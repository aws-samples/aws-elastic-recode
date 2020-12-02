#!/bin/bash
export LD_LIBRARY_PATH="/usr/local/lib:/usr/local/lib64:/usr/lib:/usr/lib64:/lib:/lib64"
export PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

#INPUT_VIDEO_S3_URL="s3://ffmpeg.zhixue/main.mp4" # S3 url of the video to test
#ORIGIN_VIDEO_S3_URL="s3://ffmpeg.zhixue/ref.mp4"   # S3 url of the reference video
#S3_BUCKET_URL="s3://ffmpeg.zhixue/logs/"        # S3 url of the log files to upload
#INPUT_VIDEO_FILE="main.mp4"  # file name of the video to test
#ORIGIN_VIDEO_FILE="ref.mp4"    # file name of the reference video

echo "CHECK ENV"
if [ -z "$INPUT_VIDEO_S3_URL" ]; then
    echo "Miss ENV INPUT_VIDEO_S3_URL"
    exit 1
fi

if [ -z "$ORIGIN_VIDEO_S3_URL" ]; then
    echo "Miss ENV $ORIGIN_VIDEO_S3_URL"
    exit 1
fi

if [ -z "$S3_BUCKET_URL" ]; then
    echo "Miss ENV S3_BUCKET_URL"
    exit 1
fi

if [[ "$S3_BUCKET_URL" == */ ]]
then
    echo "${S3_BUCKET_URL} is prefix "
else
    S3_BUCKET_URL="${S3_BUCKET_URL}/"
    echo "${S3_BUCKET_URL} add / "
fi


if [ -z "$SCALE" ]; then
    SCALE="1920x1080"
fi

if [ -z "$PSNR" ]; then
    PSNR="disable"
fi

if [ -z "$SSIM" ]; then
    SSIM="disable"
fi

if [ -z "$MSSSIM" ]; then
    MSSSIM="disable"
fi

logfile="`pwd`/video/log/${INPUT_VIDEO_FILE}.log"
logfileJsonResult="`pwd`/video/log/${INPUT_VIDEO_FILE}.json"
vmafJsonResult="`pwd`/video/log/${INPUT_VIDEO_FILE}_vmaf.json"
ffmpeg_filter_param="[0:v]scale=${SCALE}:flags=bicubic[main];[main][1:v]libvmaf=log_fmt=json:log_path=${vmafJsonResult}"

#vmaf_process
vmaf_process() {
    
    if [ "$PSNR" = "enable" ];then
        ffmpeg_filter_param=${ffmpeg_filter_param}":psnr=1"
    fi
    
    if [ "$SSIM" = "enable" ];then
        ffmpeg_filter_param=${ffmpeg_filter_param}":ssim=1"
    fi
    
    if [ "$MSSSIM" = "enable" ];then
        ffmpeg_filter_param=${ffmpeg_filter_param}":ms_ssim=1"
    fi
    
    echo "Start to run ffmpeg-vmaf command: ${ffmpeg_filter_param}"
    #create test script for debug
    #echo "ffmpeg -i ${testFile} -i ${refFile} -filter_complex \"${ffmpeg_filter_param}\" -f null -" > vmaf_test.sh
    #chmod 775 ./vmaf_test.sh
   
    (time ffmpeg -i ${testFile} -i ${refFile} -filter_complex "${ffmpeg_filter_param}" -f null - \
    2>&1) > ${logfile} 2>&1

    # get vmaf speed and metrics scores
    speed=`cat ${logfile} | awk -F'=' '/speed/ {print $NF}' | tr -d ' '`
    VMAF=`cat ${vmafJsonResult} | awk -F':' '/"VMAF score"/ {print $NF}' | tr -d ','`

    if [ "$PSNR" = "enable" ];then
        PSNR=`cat ${vmafJsonResult} | awk -F':' '/"PSNR score"/ {print $NF}' | tr -d ','`
    fi
    
    if [ "$SSIM" = "enable" ];then
        SSIM=`cat ${vmafJsonResult} | awk -F':' '/"SSIM score"/ {print $NF}' | tr -d ','`
    fi
    
    if [ "$MSSSIM" = "enable" ];then
        MSSSIM=`cat ${vmafJsonResult} | awk -F':' '/"MS-SSIM score"/ {print $NF}' | tr -d ','`
    fi
    
    # log job info
    echo "{\"scale\":\"${SCALE}\" , \"VMAF\": \"${VMAF}\", \"PSNR\": \"${PSNR}\" , \"SSIM\":\"${SSIM}\", \"MS-SSIM\":\"${MSSSIM}\", \"speed\":\"${speed}\"}" |  tr "\\n" " " > $logfileJsonResult
    
}

echo "Step1, AWS S3 copy test video ${INPUT_VIDEO_S3_URL} and reference video ${ORIGIN_VIDEO_S3_URL} to local"

mkdir -p `pwd`/video/main `pwd`/video/ref `pwd`/video/log

testFile="`pwd`/video/main/${INPUT_VIDEO_FILE}"
refFile="`pwd`/video/ref/${ORIGIN_VIDEO_FILE}"

if [ -f "$testFile" ]; then
    echo "$testFile exist, don't copy it"

else
    echo "$testFile does not exist, copy it to local"
    aws s3 cp "${INPUT_VIDEO_S3_URL}" `pwd`/video/main
fi

if [ -f "$refFile" ]; then
    echo "$refFile exist, don't copy it"

else
    echo "$refFile does not exist, copy it to local"
    aws s3 cp "${ORIGIN_VIDEO_S3_URL}" `pwd`/video/ref
fi

echo "Step2, vmaf job starting"

vmaf_process

echo "Step3, AWS S3 copy log files to ${S3_BUCKET_URL}"

aws s3 cp $logfile  ${S3_BUCKET_URL}
aws s3 cp $logfileJsonResult  ${S3_BUCKET_URL}
aws s3 cp $vmafJsonResult  ${S3_BUCKET_URL}

echo "Step4, Clear video ,log"
#remote all video , log 
rm -rf `pwd`/video/*