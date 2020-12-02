##sqs_worker环境配置
1. sqs_worker工作流程
   * 订阅SQS,消费job
   * 接受到job消息, 通过aws cli复制到本地
   * 使用ffmpeg 转吗
   * 通过aws cli sync 复制到目标s3桶，并清理资源
     

2. 配置开发环境

   ```bash
    git clone https://code.awsrun.com/jasonxue/ElasticRecodeLab.git
    python3 -m venv sqs_worker/.env
    cd sqs_worker/
    source .env/bin/activate

     #更新pip , 安装boto3
     pip3 install --upgrade pip
     pip3 intall boto3
   ```

3. CPU 版本Docker image 打包

   ```bash
   #构建docker image ,请更换成你自己的tag
   docker build -t stevensu/sqs-video-worker .
   ```
  ```

4. 本地运行

​```bash
#创建sqs队列
$ export JOB_QUEUE=ffmpeg_job_test
$ aws sqs create-queue --queue-name $JOB_QUEUE --region us-east-2

#使用stevensu/sqs-video-worker
docker pull stevensu/sqs-video-worker

#本地docker运行需要设置对应环境变量,并且需要对s3,sqs有访问权限
#环境变量说明说明
#WORKER_TYPE ffmpeg|debug, debug 仅仅打印输出
#CORE CPU|GPU 默认为CPU
#QUEUE_NAME job订阅的队列
#CONTROL_PLANE 控制队列默认是control_plane
#AWS_DEFAULT_REGION SQS队列所在的region
#AWS_ACCESS_KEY_ID Access Key
#AWS_SECRET_ACCESS_KEY=<SECRET_ACCESS_KEY>

docker run --env WORKER_TYPE=ffmpeg --env QUEUE_NAME=$JOB_QUEUE --env CORE=CPU --env AWS_DEFAULT_REGION=us-east-1  --env AWS_ACCESS_KEY_ID=<ACCESS_KEY> --env AWS_SECRET_ACCESS_KEY=<SECRET_ACCESS_KEY>   -d  stevensu/sqs-video-worker



#发送测试job,请替换input,output为你自己的s3桶和视频文件地址
#scale 1920:1080|1280:720|800:480|640:360|480:270
#profile speed|quality
Please refer to [job.json](./job.json)

#发送job.json,请确认配置了开发环境
python3 sqs_batch_tasks_producer.py -q $JOB_QUEUE -r us-east-2


#输出:
send message:
 {"userid": "8242d788-3577-455b-9927-12fa48c52fe7", "jobid": "d8356414-53e4-4357-a1ba-46d6be3f18b5", "input": "s3://m.azeroth.one/static/video/example.mp4", "output": "s3://m.azeroth.one/static/video/", "profile": {"ffmpeg": {"codec": "libx264", "scale": "480:270", "bitrate": "1M", "buffersize": "40k", "profile": "quality"}}}
 
  ```

```bash
#查看worker输出
docker logs -f $(docker ps | head -n 2| grep /app/sqs_comsumer.py|awk '{print $1}')
```


>worker 输出示例
```
Step1, AWS S3 copy  s3://m.azeroth.one/static/video/example.mp4 to local
download: s3://m.azeroth.one/static/video/example.mp4 to ../video/example.mp4
Step2, ffmpeg CPU job starting , /video/example.mp4
ffmpeg -i /video/example.mp4 -c:v libx264 -b:v 1M -bufsize 40k -vf scale=480:270 -c:a copy -profile:v high -tune psnr -threads 2 -vsync 0 /video/270/example.mp4
ffmpeg version 4.2.1 Copyright (c) 2000-2019 the FFmpeg developers
  built with gcc 9.2.0 (Alpine 9.2.0)
  configuration: --prefix=/usr --enable-avresample --enable-avfilter --enable-gnutls --enable-gpl --enable-libass --enable-libmp3lame --enable-libvorbis --enable-libvpx --enable-libxvid --enable-libx264 --enable-libx265 --enable-libtheora --enable-libv4l2 --enable-postproc --enable-pic --enable-pthreads --enable-shared --enable-libxcb --disable-stripping --disable-static --disable-librtmp --enable-vaapi --enable-vdpau --enable-libopus --disable-debug
  libavutil      56. 31.100 / 56. 31.100
  libavcodec     58. 54.100 / 58. 54.100
  libavformat    58. 29.100 / 58. 29.100
  libavdevice    58.  8.100 / 58.  8.100
  libavfilter     7. 57.100 /  7. 57.100
  libavresample   4.  0.  0 /  4.  0.  0
  libswscale      5.  5.100 /  5.  5.100
  libswresample   3.  5.100 /  3.  5.100
  libpostproc    55.  5.100 / 55.  5.100
Input #0, mov,mp4,m4a,3gp,3g2,mj2, from '/video/example.mp4':
  Metadata:
    major_brand     : mp42
    minor_version   : 0
    compatible_brands: mp42mp41isomavc1
    creation_time   : 2015-08-07T09:13:32.000000Z
  Duration: 00:00:30.53, start: 0.000000, bitrate: 2578 kb/s
   ....................
Step3, AWS S3 copy new video to s3://m.azeroth.one/static/video/270
upload: ../video/270/example.mp4 to s3://m.azeroth.one/static/video/270/example.mp4
```

