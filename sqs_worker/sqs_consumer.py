#!/usr/bin/env python3
# -*-coding:UTF-8-*-

#author  suwei007@gmail.com

from common import *
from common import recode_constant

# import os
# import os.path
# import subprocess
# import json
# import sys
import traceback

from datetime import datetime
from enum import Enum

# import boto3


WORKERS = ["ffmpeg", "vmaf","debug"]

ACTION = ["WorkerStart", "JobStart", "JobFinished", "JobError"]


class WorkerAction(Enum):
    """this is enum job status class"""
    WorkerStart = 0
    JobStart = 1
    JobFinished = 2
    JobError = 3


def timestamp():
    """timestamp helper func"""
    now = datetime.now()
    return int(datetime.timestamp(now))


#JobWorker is sqs_comsumer
class JobWorker(object):
    """This is jobworker , comsummer sqs and process ffmpeg"""

    # pylint: disable=too-many-instance-attributes
    # Eight is reasonable in this case.

    def __init__(self, worker_type='ffmpeg'):
        self.worker_type = worker_type
        self.queue_name = os.environ.get('QUEUE_NAME', 'ffmpeg_job')
        self.core = os.environ.get('CORE', 'CPU')
        self.control_plane_queue_name = os.environ.get(
            'CONTROL_PLANE', 'control_plane')
        self.ffmpeg_script = os.environ.get('FFMPEG_SCRIPT', '/app/'+self.worker_type+'_job.sh')

        self.sqs = boto3.resource('sqs')
        self.queue = self.sqs.get_queue_by_name(QueueName=self.queue_name)
        self.control_plane_queue = self.sqs.get_queue_by_name(
            QueueName=self.control_plane_queue_name)

    def start(self):
        """start comsumer message from SQS """
        print("AWS SQS Comsumer start complete at {0}".format(timestamp()))
        print("Worker [{0}] Info: CONTROL_PLANE_QUEUE:[{1}], JOB_QUEUE:[{2}]".format(
            self.worker_type, self.control_plane_queue_name, self.queue_name))
        self.send_message_to_control_plan(WorkerAction.WorkerStart, job={})

        while True:
            messages = self.queue.receive_messages(WaitTimeSeconds=5)
            for message in messages:
                print("Worker[{0}][{1}] Message received: {2}".format(self.worker_type, self.core, message.body))
                try:
                    # if self.worker_type == "ffmpeg":
                    data = json.loads(message.body)
                    # prepare env
                    env =None
                    if self.worker_type=='ffmpeg':
                        env = self.build_ffmpeg_env(data)
                    if self.worker_type=='vmaf':
                        env = self.build_vmaf_env(data)

                    if env is None:
                        self.send_message_to_control_plan(
                            action=WorkerAction.JobError, job=data,msg='build env failed')
                        message.delete()
                        continue
                    #send JobStart to controlplane
                    self.send_message_to_control_plan(
                        action=WorkerAction.JobStart, job=data)
                    #process job when job complete delete message from SQS
                    subprocess.call([self.ffmpeg_script], env=env)
                    message.delete() # here we may meet problems, we can only delete the message while the transcode job is completed, or it can be controlled outside from the user side to resubmit failed jobs
                    
                    #send JobFinished to controlplane
                    self.send_message_to_control_plan(
                        action=WorkerAction.JobFinished, job=data)
                except Exception as e:
                    traceback.print_exc()
                    self.send_message_to_control_plan(action=WorkerAction.JobError, job=data, msg=e)
                    print("Error", e)

    def send_message_to_control_plan(self, action, job, msg=''):
        """send message to control plan queue """
        data = {}
        data["workerType"] = self.worker_type
        data["nodeName"] = os.environ.get('WORKER_NODE_NAME', 'UNKNOW')
        data["podName"] = os.environ.get('WORKER_POD_NAME', 'UNKNOW')
        data["action"] = ACTION[action.value]
        data["job"] = job
        data["msg"] = msg
        data["ts"] = timestamp()
        body = json.dumps(data)
        self.control_plane_queue.send_message(MessageBody=body)

    def build_vmaf_env(self, data):
        try:
            print(data)
            _, inputFile = os.path.split(data['input'])
            _, origin = os.path.split(data['origin'])
            env = {
                'INPUT_VIDEO_S3_URL': data['input'],
                'INPUT_VIDEO_FILE': inputFile,
                'ORIGIN_VIDEO_S3_URL': data['origin'],
                'ORIGIN_VIDEO_FILE': origin,
                'SCALE': data['profile']['vmaf']['scale'],
                'SSIM':data['profile']['vmaf']['ssim'],
                'PSNR':data['profile']['vmaf']['psnr'],
                'MSSSIM':data['profile']['vmaf']['ms-ssim'],
                'S3_BUCKET_URL': data['output'],
                'AWS_DEFAULT_REGION': os.environ.get('AWS_DEFAULT_REGION', 'us-east-1'),
                'AWS_ROLE_ARN': os.environ.get('AWS_ROLE_ARN', ''),
                'AWS_ACCESS_KEY_ID':os.environ.get('AWS_ACCESS_KEY_ID', ''),
                'AWS_SECRET_ACCESS_KEY':os.environ.get('AWS_SECRET_ACCESS_KEY', ''),
                'AWS_WEB_IDENTITY_TOKEN_FILE': os.environ.get('AWS_WEB_IDENTITY_TOKEN_FILE', ''),
                'MODE': self.worker_type
            }
            return env
        except Exception:
            traceback.print_exc()
            return None

    def build_ffmpeg_env(self, data):
        """preprae env from data"""
        # pylint: disable=broad-except
        try:
            _, videofile = os.path.split(data['input'])
            scale = data['profile']['ffmpeg']['scale']
            codes = data['profile']['ffmpeg']['codec']
            env = {
                'VIDEO_S3_URL': data['input'],
                'VIDEO_FILE': videofile,
                'S3_BUCKET_URL': data['output'],
                'AWS_DEFAULT_REGION': os.environ.get('AWS_DEFAULT_REGION', 'us-east-1'),
                'AWS_ROLE_ARN': os.environ.get('AWS_ROLE_ARN', ''),
                'AWS_ACCESS_KEY_ID':os.environ.get('AWS_ACCESS_KEY_ID', ''),
                'AWS_SECRET_ACCESS_KEY':os.environ.get('AWS_SECRET_ACCESS_KEY', ''),
                'AWS_WEB_IDENTITY_TOKEN_FILE': os.environ.get('AWS_WEB_IDENTITY_TOKEN_FILE', ''),
                'CORE': self.core,
                'CODEC': codes,
                'BITRATE': data['profile']['ffmpeg']['bitrate'],
                'BUFFERSIZE': data['profile']['ffmpeg']['buffersize'],
                'SCALE': scale,
                'W': scale.split(":")[0],
                'H': scale.split(":")[1],
                'MODE': self.worker_type
            }

            if data['profile']["ffmpeg"]["profile"] == recode_constant.USER_PROFILE_LATENCY:
                env["SPEED_PROFILE"] = "true"

            if self.core == "GPU":
                env["ORIGIN_CODEC"] = data['profile']['ffmpeg']['originCodec']

            return env
        except Exception:
            traceback.print_exc()
            return None


if __name__ == "__main__":
    # pylint: disable=broad-except,invalid-name
    try:
        worker_type = os.environ.get('WORKER_TYPE', 'ffmpeg_job')
        if worker_type not in WORKERS:
            print("Worker type not support {0}".format(worker_type))
            sys.exit()
        print('init Q sub')
        JobWorker(worker_type).start()
    except Exception as e:
        print(e)
        sys.exit()
