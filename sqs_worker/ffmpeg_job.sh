#!/bin/bash
export LD_LIBRARY_PATH="/usr/local/lib:/usr/local/lib64:/usr/lib:/usr/lib64:/lib:/lib64"
export PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

echo "CHECK ENV"
if [ -z "$VIDEO_S3_URL" ]; then
    echo "Miss ENV VIDEO_S3_URL"
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
    SCALE="800:480"
fi

if [ -z "$CODEC" ]; then
    SCALE="libx264"
fi

if [ -z "$BITRATE" ]; then
    BITRATE="1M"
fi

if [ -z "$BUFFERSIZE" ]; then
    BUFFERSIZE="2M"
fi

if [ -z "$THREADS" ]; then
    THREADS="2"
fi


if [ -z "$MODE" ]; then
    MODE="debug"
fi


SCALE_PREFIX=${SCALE##*:}
logfile="`pwd`/video/${SCALE_PREFIX}/${VIDEO_FILE}.log"
logfileJsonResult="`pwd`/video/${SCALE_PREFIX}/${VIDEO_FILE}.json"

#gpu process
gpu_ffmpeg_process() {
    ffmpegProfile="high"
    dec_codec="h264_cuvid"
    enc_codec="h264_nvenc"
    ffmpeg_para=""
    userProfile="quality"
    

    if [ "${CODEC}" = "libx265" ]; then
        ffmpegProfile="main"
        dec_codec="hevc_cuvid"
        enc_codec="hevc_nvenc"
    fi

    if [ -n "$SPEED_PROFILE" ];then
       echo "Latency priority"
       userProfile="latency"
       ffmpeg_para="-preset llhp -b:v ${BITRATE} -bufsize ${BUFFERSIZE} -rc cbr_ld_hq -g 999999"
    else
       echo "Quality priority"
       ffmpeg_para="-preset medium -b:v ${BITRATE} -bufsize ${BUFFERSIZE} -bf 3 -b_ref_mode 2 -temporal-aq 1 -rc-lookahead 20"

    fi
    
    echo "nohup ffmpeg -hwaccel cuvid -c:v $dec_codec -i `pwd`/video/"${VIDEO_FILE}" -c:v $enc_codec -c:a copy -profile:v $ffmpegProfile -filter:v scale_npp=w=${width[$i]}:h=${height[$i]} $ffmpeg_para -vsync 0 -stats -y `pwd`/video/${SCALE_PREFIX}/${VIDEO_FILE} 2>&1 > $logfile & "
    if [ "$MODE" = "ffmpeg" ]; then
        echo "Start to run ffmpeg command:"
        scaleArray=($(echo $SCALE | tr ":" "\n"))
        # Recode with codec h264 or 265
        nohup ffmpeg -hwaccel cuvid -c:v $dec_codec -i `pwd`/video/"${VIDEO_FILE}" -c:v $enc_codec -c:a copy -profile:v $ffmpegProfile \
            -filter:v scale_npp=w=${scaleArray[0]}:h=${scaleArray[1]} $ffmpeg_para  \
            -vsync 0 -stats -y `pwd`/video/${SCALE_PREFIX}/${VIDEO_FILE} \
            2>&1 > $logfile &

        # Get GPU utilization
        gpu_usage=`nvidia-smi dmon -c 2 | tail -n 1 | awk '{print "gpu_enc "$7" gpu_dec "$8}'`

        # Get ffmpeg speed after ffmpeg is done
        ps=`ps aux | grep -v grep | grep ffmpeg`
                
        while [[ -n "$ps" ]]
        do
           sleep 10
           ps=`ps aux | grep -v grep | grep ffmpeg`
        done
        
        speed=`cat $logfile | awk -F'=' '/speed/ {print $NF}'`
        
        # Log gpu_usage and speed
        echo "{\"codex\":\"${CODEC}\",\"scale\":\"${SCALE}\",\"UserProfile\": \"${userProfile}\",\"gpu_usage\":\"$gpu_usage\", \"speed\":\"${speed}\"}" >> $logfileJsonResult
         
    fi    
}

#cpu process
cpu_ffmpeg_process() {
    ffmpegProfile="high"
    ffmpeg_para=""
    userProfile="quality"
    if [ "${CODEC}" = "libx265" ]; then
        ffmpegProfile="main"
    fi

    if [ -n "$SPEED_PROFILE" ];then
       echo "Latency priority"
       userProfile="latency"
       ffmpeg_para="-preset fast -b:v ${BITRATE} -bufsize ${BUFFERSIZE} -g 999999 -x264opts no-sliced-threads:nal-hrd=cbr -tune psnr"

    else
       echo "Quality priority"
       ffmpeg_para="-preset medium -b:v ${BITRATE} -bufsize ${BUFFERSIZE} -tune psnr"
    fi
    
    echo "(ffmpeg -i `pwd`/video/"${VIDEO_FILE}" -c:v ${CODEC} -c:a copy -vf scale=${SCALE} -profile:v $ffmpegProfile $ffmpeg_para -vsync 0 -v quiet -stats -y `pwd`/video/${SCALE_PREFIX}/${VIDEO_FILE} 2>&1  ) > $logfile 2>&1 "
    if [ "$MODE" = "ffmpeg" ]; then
        echo "Start to run ffmpeg command:"
        
        (time ffmpeg -i `pwd`/video/"${VIDEO_FILE}" -c:v ${CODEC} -c:a copy -vf scale=${SCALE} \
            -profile:v $ffmpegProfile $ffmpeg_para \
            -vsync 0 -stats -y `pwd`/video/${SCALE_PREFIX}/${VIDEO_FILE} \
            2>&1 ) > $logfile 2>&1   
            
        # Get ffmpeg speed 
        speed=`cat $logfile | awk -F'=' '/speed/ {print $NF}'`
        # Get CPU usage
        user=`cat $logfile | awk '/user/ {print $2}' | awk -F'm' '{printf 60*int($1*100)/100+int($2*100)/100}'`
        real=`cat $logfile | awk '/real/ {print $2}' | awk -F'm' '{printf 60*int($1*100)/100+int($2*100)/100}'`
        # cpu_usage=`echo "scale=4;$user/$real/$vcpu/2*100;" | bc`
        # Log cpu_usage and speed
        echo "{\"codex\":\"${CODEC}\",\"scale\":\"${SCALE}\",\"UserProfile\": \"${userProfile}\",\"cpu_usage\":{ \"user\":$user ,\"real\":$real}, \"speed\":\"${speed}\"}" >> $logfileJsonResult
         
    fi
}

echo "env:"  ${VIDEO_S3_URL} ${S3_BUCKET_URL} ${SCALE}, ${SCALE_PREFIX}

echo "Step1, AWS S3 copy  ${VIDEO_S3_URL} to local"

mkdir -p `pwd`/video/${SCALE_PREFIX}

inputFile="`pwd`/video/${VIDEO_FILE}"

if [ -f "$inputFile" ]; then
    echo "$inputFile exist, don't copy it"

else
    echo "$inputFile does not exist, copy it to local"
    aws s3 cp "${VIDEO_S3_URL}" `pwd`/video
fi

echo "Step2, ffmpeg ${CORE} job starting , `pwd`/video/${VIDEO_FILE}"

if [ "$CORE" = "GPU" ]; then
   gpu_ffmpeg_process
else
   cpu_ffmpeg_process
fi

echo "Step3, AWS S3 copy new video `pwd`/video/${SCALE_PREFIX}/${VIDEO_FILE} to ${S3_BUCKET_URL}"

if [ "$MODE" = "ffmpeg" ];then
    aws s3 cp `pwd`/video/${SCALE_PREFIX}/${VIDEO_FILE}  ${S3_BUCKET_URL}
    aws s3 cp $logfile  ${S3_BUCKET_URL}${VIDEO_FILE}.log
    aws s3 cp $logfileJsonResult  ${S3_BUCKET_URL}${VIDEO_FILE}.json
fi

rm -rf `pwd`/video/*

