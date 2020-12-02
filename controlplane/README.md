
##AWS ElasticRecode Control Plane 部署

###v0.1.4 版本说明

* 增加VMAF API, 详情请见**3.API说明中 /api/v1/vmaf** 

* 自动发现worker 支持worker类型: ffmpeg(转码),vmaf(质量评价)

* 首页自动获取版本

* 增加buildtime和gitcommit log(示例如下)

   ```bah
  version: v0.1.4
  build: 2020.05.17.103632
   ```

  

###v0.1.3 版本说明

* k8s 部署请使用 stevensu/elasticrecode_control_plane:k8s_v0.1.3
* 本地docker 部署请使用 stevensu/elasticrecode_control_plane:v0.1.3

* 增加batch job处理,批量任务仅支持同一个profile, output(prefix), 具体参考API说明
* 提交job返回内容修改,不再返回具体的profie, 将返回事件状态, 具体参考API说明

### v0.1.2 版本说明

* 增加支持多job队列, 当同一平台(CPU|GPU)有多个队列时候，系统会随机发送到某个队列,未来会配合autodiscovery功能增加自动权重

* 增加dynamodb 存储日志 , 禁止该功能 -disableWriteJobLogs

* 增加K8S deployment 自动发现功能 禁止该功能 -disableAutoDiscovery

  AutoDiscovery 功能可以自动发现有特定annotation的deployment,自动注册它们使用的job队列,并统计它的副本数量, 该功能需要worker deployment的部署文件使用特定的annotation ，并且worker环境变量CONTROL_PLANE设置和controlplane的控制队列相同. 多个worker deployment可以配合节点亲和策略调度到不同的EC2上去，它们可以共同使用一个job队列也可以使用多个job队列

  * annotation说明

    ```yaml
    elastirecord.kubernetes.io/worker: 'enable'
    elastirecord.kubernetes.io/platform: 'cpu|gpu'  
    ```

  

* 增加API

  * schema 

  * 创建 job

  * job 日志 , 支持单个job,某个用户所有job的查询 (-disableWriteJobLogs 启用时该API无效)

  * worker queue 查询

  * worker deployment 查询(仅autodiscovery启动时生效)

    

  

  

###1. 本地docker运行

​    stevensu/elasticrecode_controlplane:v0.1.2  后台没有运行守护进程,方便本地测试

前提条件

* 正在运行的worker,本地/K8S均可 **(必须)**

* SQS 控制队列 **(必须)**
  
  ```bash
    #创建控制队列,如果测试worker的时候已经创建,请跳过
    aws sqs create-queue --queue-name control_plane
    
    #创建worker工作队列,如果测试worker的时候已经创建请跳过
  aws sqs create-queue --queue-name ffmpeg_job
    
    #运行控制平面,详细参数见后面说明
    docker run --rm -p 8080:8080 --env QUEUE_NAME=control_plane --env AWS_DEFAULT_REGION=us-east-1  --env AWS_SECRET_ACCESS_KEY=<替换SecretAccessKey> --env AWS_ACCESS_KEY_ID=<替换AccessKey> stevensu/elasticrecode_control_plane:v0.1.3 /app/controlplane -q control_plane --disableAutoDiscovery -disableWriteJobLogs --cpuQ=ffmpeg_job
    
   #打开浏览器
   http://localhost:8080/
   
   #发送任务
   #请修改job.json里面的input/output为worker可以访问的s3地址
   curl -v -X POST -H "Content-Type: application/json" --data @job_from_ui.json localhost:8080/api/v1/jobs
   
  ```
  
  
  
* 使用DynamoDB存储job日志  (可选, 可以使用-disableWriteJobLogs 禁止使用dynamodb 存储job日志)
  
  **注意:** dynamodb同时只能创建1个GSI索引，所以请再控制台确认创建成功后再创建第2个
  
  ```bash
    #创建dynamodb
    aws dynamodb create-table \
     --table-name elasticrecode_logs \
     --attribute-definitions AttributeName=ts,AttributeType=N \
     --key-schema AttributeName=ts,KeyType=HASH  \
     --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 
     
    #创建GSI索引 userIDIndex
    aws dynamodb update-table \
    --table-name elasticrecode_logs \
    --attribute-definitions \
  AttributeName=userID,AttributeType=S \
      AttributeName=ts,AttributeType=N \
    --global-secondary-index-updates \
  "[{\"Create\":{\"IndexName\": \"userIDIndex\",\"KeySchema\":[{\"AttributeName\":\"userID\",\"KeyType\":\"HASH\"},{\"AttributeName\":\"ts\",\"KeyType\":\"RANGE\"}], \
  \"ProvisionedThroughput\": {\"ReadCapacityUnits\": 10, \"WriteCapacityUnits\": 5      },\"Projection\":{\"ProjectionType\":\"ALL\"}}}]"
  
  
   #创建GSI索引 jobIDIndex
   aws dynamodb update-table \
        --table-name elasticrecode_logs \
        --attribute-definitions \
          AttributeName=jobID,AttributeType=S \
          AttributeName=ts,AttributeType=N \
        --global-secondary-index-updates \
        "[{\"Create\":{\"IndexName\": \"jobIDIndex\",\"KeySchema\":[{\"AttributeName\":\"jobID\",\"KeyType\":\"HASH\"},{\"AttributeName\":\"ts\",\"KeyType\":\"RANGE\"}], \
      \"ProvisionedThroughput\": {\"ReadCapacityUnits\": 10, \"WriteCapacityUnits\": 5      },\"Projection\":{\"ProjectionType\":\"ALL\"}}}]"
      
  
   #取消-disableWriteJobLogs参数,重新运行控制平面
    docker run --rm -p 8080:8080 --env QUEUE_NAME=control_plane --env AWS_DEFAULT_REGION=us-east-1  --env AWS_SECRET_ACCESS_KEY=<替换SecretAccessKey> --env AWS_ACCESS_KEY_ID=<替换AccessKey> stevensu/elasticrecode_controlplane:v0.1.2 /app/controlplane -q control_plane --disableAutoDiscovery --cpuQ=ffmpeg_job
    
  #使用浏览器或者curl 访问 http://localhost:8080/api/v1/logs/{job.json里面的userid}
  http://localhost:8080/api/v1/logs/8242d788-3577-455b-9927-12fa48c52fe7
  
  ```
* 清除环境

  ```bash
      
  #清除控制队列
  aws sqs delete-queue --queue-url $(aws sqs get-queue-url --queue-name control_plane --out text)
  #清除dynamodb(可选)
  #aws dynamodb delete-table --table-name  elasticrecode_logs
  ```



###2. 部署到K8S



  Control Plane 可以自动发现K8S集群内的使用了annotation的deployment.

  认证方式有2种:

​    1). incluster, 部署到k8s集群内，通过service account授权.

​    2). kubeconfig 远程访问k8s集群模式（因为是远程获取deployment事件所以不推荐这种方式,仅供开发使用)

禁止访问k8s,使用以下参数就可以关闭K8S集群访问

  -disableAutoDiscovery 



* InCluster 模式

利用eksctl 创建service account,配置IAM 权限为控制平面创建IRSA权限,为了方便测试启用了SQS,DynamoDB,EC2,CloudWatch的权限

```bash
  #控制平面创建IRSA权限,为了方便测试启用了SQS,DynamoDB,EC2,CloudWatch的权限
  export EKSCLUSTER=eksworkshop
eksctl create iamserviceaccount  \
            --name elasticrecode-controlplane \
             --cluster $EKSCLUSTER \
             --attach-policy-arn arn:aws:iam::aws:policy/AmazonSQSFullAccess \
             --attach-policy-arn arn:aws:iam::aws:policy/AmazonDynamoDBFullAccess \
             --attach-policy-arn arn:aws:iam::aws:policy/CloudWatchFullAccess \
             --attach-policy-arn arn:aws:iam::aws:policy/AmazonEC2FullAccess --approve
  
  #为service account 设置RABC权限
  kubectl apply -f controlplane-rabc.yaml
  
  #部署controlplane到eks集群
  kubectl apply -f controlplane.yaml
  
  #使用带有annotation的worker
  annotations:
    elastirecord.kubernetes.io/worker: 'enable'
    elastirecord.kubernetes.io/platform: 'cpu'
  
```



###3. API说明

 测试controlplane地址:

 http://a3d7207e28fac11eaaea312e1b79311b-65109e6ee5950866.elb.us-east-1.amazonaws.com/

GET Method:

* /api/v1/schema 基础参数,数组表示返回取值， 返回

  ```json
  {"ec2":{"platform":["cpu","gpu"],"priceModel":["normal","spot"]},"ffmpeg":{"codec":["h264","h265"],"originCodec":["h264","h265"],"scale":["720p","480p","360p","240p","1080p"],"bitrate":[1,15],"profile":["quality","latency"],"priority":[0,100]}}
  ```

  

* /api/v1/jobs/{jobID}  具体job的信息,返回示例 

  

  ```json
  [
    {
      "ts": 1588749694,
      "userID": "8242d788-3577-455b-9927-12fa48c52fe7",
      "jobID": "460c8ab9-4741-49b2-bf9f-f97b87e5efc6",
      "action": "JobStart",
      "job": {
        "userid": "8242d788-3577-455b-9927-12fa48c52fe7",
        "jobid": "460c8ab9-4741-49b2-bf9f-f97b87e5efc6",
        "input": "s3://m.azeroth.one/static/video/example.mp4",
        "output": "s3://m.azeroth.one/static/video/output/",
        "priority": 100,
        "profile": {
          "ffmpeg": {
            "codec": "libx264",
            "scale": "1920:1080",
            "bitrate": "1M",
            "buffersize": "2M",
            "profile": "quality",
            "platform": "cpu"
          }
        }
      },
      "nodeName": "ip-192-168-63-212.ec2.internal",
      "podName": "ffmpeg-worker-95d8865bb-gtpnv"
    },
    {
      "ts": 1588749739,
      "userID": "8242d788-3577-455b-9927-12fa48c52fe7",
      "jobID": "460c8ab9-4741-49b2-bf9f-f97b87e5efc6",
      "action": "JobFinished",
      "job": {
        "userid": "8242d788-3577-455b-9927-12fa48c52fe7",
        "jobid": "460c8ab9-4741-49b2-bf9f-f97b87e5efc6",
        "input": "s3://m.azeroth.one/static/video/example.mp4",
        "output": "s3://m.azeroth.one/static/video/output/",
        "priority": 100,
        "profile": {
          "ffmpeg": {
            "codec": "libx264",
            "scale": "1920:1080",
            "bitrate": "1M",
            "buffersize": "2M",
            "profile": "quality",
            "platform": "cpu"
          }
        }
      },
      "nodeName": "ip-192-168-63-212.ec2.internal",
      "podName": "ffmpeg-worker-95d8865bb-gtpnv"
    }
  ]
  ```

  

* /api/v1/{userID}/jobs 某个用户的所有job,返回示例

  

* /api/v1/worker/queues 当前worker队列,返回示例

* /api/v1/worker/deployments K8S自动发现模式下的deployment信息, 返回示例

  ```json
  {"ffmpeg-worker":{"Name":"ffmpeg-worker","Queue":"ffmpeg_job","Platform":"cpu","Replicas":1}}
  ```

  

POST Method:

* /api/v1/vmaf 发布视频质量评估任务到controlplane , 仅支持单个job，参数如下

  * **output 字段均统一为prefix**,自动会自动根据input的文件名字设置

  * input 字段为需要评测文件

  * origin 字段为原始文件(未转码文件)

    示例如下:

    ```json
    {
        "userid":"8242d788-3577-455b-9927-12fa48c52fe7",
        "input":"s3://m.azeroth.one/static/video/example.mp4",
        "origin":"s3://m.azeroth.one/static/video/example.mp4",
        "output":"s3://m.azeroth.one/static/video/vmaf/",
        "profile":{
        "vmaf":{
           "scale":"720p",
           "ssim":"enable",
           "psnr":"enable",
           "ms-ssim":"enable"
         }
        }
        
    }
    ```

    返回内容(会带有jobid和status，后续需要通过/api/v1/jobs/{jobid} 去访问最终Job状态):

    ```json
    {
      "userid": "8242d788-3577-455b-9927-12fa48c52fe7",
      "jobid": "547124b6-623c-4039-a740-aa5caf74c7eb",
      "input": "s3://m.azeroth.one/static/video/example.mp4",
      "origin": "s3://m.azeroth.one/static/video/example.mp4",
      "output": "s3://m.azeroth.one/static/video/vmaf/",
      "profile": {
        "vmaf": {
          "scale": "1280x720",
          "ssim": "enable",
          "psnr": "enable",
          "ms-ssim": "enable"
        }
      },
      "status": {
        "event": "JobSummit",
        "status": "success"
      }
    }
    
    ```

    

* /api/v1/jobs 发布job到controlplane , 单个/批量使用的入口一致，参数不一样
  
* **output 字段均统一为prefix**,自动会自动根据input的文件名字设置
    
* 单个job发布, **示例请参考job.json**
  
  * 批量job发布, **批量参数使用batchInputs，类型数组 示例请参考job_batch.json**
  
    在批量job提交是请不要携带input参数,否则将视为单个任务发布.
  
    ```json
    {
    	"userid":"8242d788-3577-455b-9927-12fa48c52fe7",
    	"batchInputs":[
    		"s3://m.azeroth.one/static/video/example.mp4",
    		"s3://m.azeroth.one/static/video/example1.mp4",
    		"s3://m.azeroth.one/static/video/example2.mp4"
    		],
    	"output":"s3://m.azeroth.one/static/video/output/",
    	"priority":100,
    	"profile":{
    		"ffmpeg":{
    			"codec":"h264",
    			"scale":"1080p",
    			"bitrate":"1M",
    			"profile":"quality",
    			"platform":"cpu"
    		}
    	}
    }
  ```
  
  * 单个/批量job 返回, **返回类型均为数组**
  
    ```json
    [
      {
        "userid": "8242d788-3577-455b-9927-12fa48c52fe7",
        "jobid": "9c5ecde6-15af-4b96-a456-4556d57694e2",
        "input": "s3://m.azeroth.one/static/video/example.mp4",
        "output": "s3://m.azeroth.one/static/video/output/",
        "status": {
          "event": "JobSummit",
          "status": "success"
        }
      },
      {
        "userid": "8242d788-3577-455b-9927-12fa48c52fe7",
        "jobid": "1d31e5f3-be43-4677-93c8-b6e14af0a3b1",
        "input": "s3://m.azeroth.one/static/video/example1.mp4",
        "output": "s3://m.azeroth.one/static/video/output/",
        "status": {
          "event": "JobSummit",
          "status": "success"
        }
      },
      {
        "userid": "8242d788-3577-455b-9927-12fa48c52fe7",
        "jobid": "4ef15298-6d77-4e2d-b1f4-5d91e300f95c",
        "input": "s3://m.azeroth.one/static/video/example2.mp4",
        "output": "s3://m.azeroth.one/static/video/output/",
        "status": {
          "event": "JobSummit",
          "status": "success"
        }
    }
    ]
    ```
  
    

WebSocket Method:

  * /ws?lastMod=0 所有事件通知
  ```javascript
  //参考home.html的webocket链接
  "ws://" + document.location.host + "/ws?lastMod=0
  ```

  

###4. UI集成

controlplane  默认主页为home.html

docker image 已经内置了静态目录为/static ,运行后可以访问测试js, http://localhost:8080/static/jquery-3.5.1.min.js, 可以通过docker run -v <本地静态文件>:/app/static 挂载