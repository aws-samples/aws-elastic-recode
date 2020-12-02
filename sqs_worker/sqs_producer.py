#!/usr/bin/env python3
# -*-coding:UTF-8-*-

import os
import os.path
import subprocess
import json
import sys

from datetime import datetime
import sys, getopt

import boto3



data_file="job.json"

if len(sys.argv)==2:
    data_file=sys.argv[1]



sqs = boto3.resource('sqs')


def send_messag(queue_name,job):
    f=open(job)
    data = json.load(f)
    body = json.dumps(data)
    control_plane_queue=sqs.get_queue_by_name(QueueName=queue_name)
    response = control_plane_queue.send_message(MessageBody=body)
    print("send message:\n {0}".format(body))







def main(argv):
   queue = ''
   job = ''
   try:
      opts, args = getopt.getopt(argv,"hq:j:",["queue=","region="])
   except getopt.GetoptError:
      print ('GetoptError, usage: sqs_producer.py -q <queue_name> -j <json file>')
      sys.exit(2)
   for opt, arg in opts:
      if opt == '-h':
         print ('usage: sqs_producer.py -q <queue_name> -j <json file>')
         sys.exit()
      elif opt in ("-q", "--queue"):
         queue = arg
      elif opt in ("-j", "--job"):
         job = arg
   if queue == '' or  job== "":
       print ('usage: sqs_producer.py -q <queue_name> -f <json file>')
   else:
        send_messag(queue,job)

   
if __name__ == "__main__":
   main(sys.argv[1:])

