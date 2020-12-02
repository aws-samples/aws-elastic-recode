#!/usr/bin/env python3
# -*-coding:UTF-8-*-

#author  Jason Xue

from common import *
from common import recode_constant

## The buffersize is determined by different profile value
## This function helps to calcuate the buffersize from the input bitrate & profile & codec
def _buffersizeByProfile(bitrate,profile_name,codex=""):
   bitrate = bitrate.upper()
   buffersize = bitrate
   length = len(bitrate)
   unit = bitrate[length - 1]
   if length >= 2:
      bitrateValue = bitrate.rstrip("M")
      bitrateValue = float(bitrateValue.rstrip("K"))

      if profile_name == recode_constant.USER_PROFILE_QUALITY :
         buffersize = int(bitrateValue * 2)
      elif profile_name == recode_constant.USER_PROFILE_LATENCY :
         divisor = 25
         if codex == recode_constant.CODEC_265:
            divisor = 4
         if unit == "M":
            buffersize = int(bitrateValue * 1024 / divisor)
            unit = "K"
         else:
            buffersize = int(bitrateValue / divisor)
      return str(buffersize)+unit
   
   return buffersize
###
 ## Conver the user input options into Jobs Json results
 ## fingerprint - a random number or sequence number for different jobs
 ## inputVideo - the full path of S3 object
 ## sizeOptions - the sizes to be transcoded
###
def _userOptionsToJobs(fingerprint,inputVideo,sizeOptions,bitrateOptions,s3_bucket,output_location,cpuPlatformOptions=[recode_constant.CPU_PLATFORM_CPU,recode_constant.CPU_PLATFORM_GPU],userProfileOptions=[recode_constant.USER_PROFILE_QUALITY,recode_constant.USER_PROFILE_LATENCY],codecOptions=[recode_constant.CODEC_264,recode_constant.CODEC_265]):
   jobs = []
   sizeToBitrate = {
      "1080p":"4M",
      "720p":"2.5M",
      "480p":"900K",
      "360p":"600K"  
   }
   videoName = inputVideo[inputVideo.rindex("/")+1:len(inputVideo)]
   outputPatten = "output-{0}-{1}-{2}-{3}-"+ videoName
   count = fingerprint
   for s in sizeOptions:
      scale = recode_constant.SIZE_TO_SCALE[s]
      bitrate = sizeToBitrate[s]
      for c in codecOptions:
         for userProfile in userProfileOptions:
            buffersize = _buffersizeByProfile(bitrate,userProfile,c)
            for platform in cpuPlatformOptions:
               count = count + 1
               jobid = str(uuid.uuid3(uuid.NAMESPACE_DNS, 'elastic_recode_job'+str(count)))
               userid = str(uuid.uuid3(uuid.NAMESPACE_DNS, 'elastic_recode_user'+str(count)))
               outputVideo = s3_bucket + output_location + outputPatten.format( c , s , userProfile , platform)
               jobJson = {"userid":userid,"jobid":jobid,"input":inputVideo,"output":outputVideo,
                  "profile":{
                     "ffmpeg": {
                        "codec" : c,
                        "scale" : scale,
                        "bitrate" : bitrate,
                        "buffersize" : buffersize,
                        "profile" : userProfile,
                        "platform" : platform
                     }
                  }
               }
               print("job# {0} - {1}".format(count,jobJson))  
               jobs.append(jobJson)
   return jobs

###
## the main function to generate the benchmark jobs
###
def benchmarkBatchJobs(queue_name,s3_bucket,input_file,output_location,region_name):
   benchmarkProfiles = {
      "sizeOptions":["1080p","720p","480p","360p"],
      "bitrateOptions":["4M", "2.5M", "900K", "600K"]
      # "ffmpegProfileOptions":[FFMPEG_PROFILE_MAIN,FFMPEG_PROFILE_HIGH]
   }
   ## Benchmark Jobs 1: Transcode a series of different resolution videos from the single video source 
   ## The number of jobs for one video could be calculated by : len(benchmarkProfiles["sizeOptions"]) * len(benchmarkProfiles["codecOptions"])

   inputVideo = s3_bucket + input_file
   benchmarkPrepareJobs = _userOptionsToJobs(1,inputVideo,benchmarkProfiles["sizeOptions"],
         benchmarkProfiles["bitrateOptions"],
         s3_bucket,
         output_location,
         [recode_constant.CPU_PLATFORM_CPU],
         [recode_constant.USER_PROFILE_QUALITY])
   count = len(benchmarkPrepareJobs)
   submitJobsToQueue(queue_name,benchmarkPrepareJobs,region_name)
   ## Based on the above output videos, we start to generate the benchmark transcode jobs, with all options combines
   benchmarkJobs = []
   for rs in benchmarkPrepareJobs:
      inputVideo = rs["output"]
      benchmarkJobs2 = _userOptionsToJobs(count+1,inputVideo,benchmarkProfiles["sizeOptions"],benchmarkProfiles["bitrateOptions"],s3_bucket,output_location)
      count = count + len(benchmarkJobs2)
      benchmarkJobs.append(benchmarkJobs2)
   
   # sqs = boto3.resource('sqs',region_name=region_name)
   # control_plane_queue=sqs.get_queue_by_name(QueueName=queue_name)


###
##  Submit the Jobs to the queue
###
def submitJobsToQueue(queue_name,jobs,region="us-east-2"):
   sqs = boto3.resource('sqs',region_name=region)
   jobsq=sqs.get_queue_by_name(QueueName=queue_name)
   for job in jobs:
      response = jobsq.send_message(MessageBody=json.dumps(job))

###
##  Empty the job queue
###
def emptyJobQueue(queue_name,region="us-east-2"):
   sqs = boto3.resource('sqs',region_name=region)
   jobsq=sqs.get_queue_by_name(QueueName=queue_name)
   jobsq.purge()



def main(argv):
   queue = ''
   s3_bucket = ''
   input_file = ''
   output_location = ''
   region = ''
   try:
      opts, args = getopt.getopt(argv,"hq:b:i:o:r:",["queue=","s3_bucket=","input_file=","output_location=","region="])
   except getopt.GetoptError:
      print ('GetoptError, usage: sqs_batch_tasks_producer.py -q <queue_name> -b <s3_bucket> -i <input_file> -o <output_location> -r <region_name>')
      sys.exit(2)
   for opt, arg in opts:
      if opt == '-h':
         print ('usage: sqs_batch_tasks_producer.py -q <queue_name> -b <s3_bucket> -i <input_file> -o <output_location> -r <region_name>')
         sys.exit()
      elif opt in ("-q", "--queue"):
         queue = arg
      elif opt in ("-b", "--bucket"):
         s3_bucket = arg
      elif opt in ("-i", "--input"):
         input_file = arg
      elif opt in ("-o", "--output"):
         output_location = arg
      elif opt in ("-r", "--region"):
         region = arg
   if queue == '' or  s3_bucket == "" or input_file == "" or output_location == "" or region== "":
       print ('usage: sqs_batch_tasks_producer.py -q <queue_name> -b <s3_bucket> -i <input_file> -o <output_location> -r <region_name>')
   else:
        benchmarkBatchJobs(queue,s3_bucket,input_file,output_location,region)

   
if __name__ == "__main__":
   main(sys.argv[1:])

