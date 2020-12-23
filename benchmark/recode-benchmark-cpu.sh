#!/bin/bash

# BEFORE running this script:
# 1. Put origin videos in the same diretory with this script
# 2. Use the following filename for origin video
    ## libx264.mp4 - for 264 video
    ## libx265.mp4 - for 265 video

if [ -z $1 ]; then
    echo "Usage: Run bench.sh <result-path> "
    exit -1
fi
if [ ! -d $1 ]; then
    echo "$1 is not a directory!"
    exit -1
fi
path=${1%/}

size=(1080p 720p 480p 360p)
width=(1920 1280 800 640)
height=(1080 720 480 360)
bandwidth=(4M 2.5M 900K 600K)
buffer=(8M 5M 1800K 1200K)
#buffer_quality=(8M 5M 1800K 1200K)
#buffer_latency_264=(166K 104K 38K 25K)
#buffer_latency_265=(2M 1250K 450K 300K)
codec=(libx264 libx265)
use_profile=(quality latency)
vcpu=`cat /proc/cpuinfo | awk -F':' '/cpu cores/ {print $NF}' | tail -n 1`


# Prepare for input files of different resolutions
for ((k=0;k<${#codec[@]};k++)); do
    # For different codec
    echo "====== Transcoding Input Files ========="
    for ((i=0;i<${#size[@]};i++)); do
        if [ "${codec[$k]}" = "libx264" ]; then
            profile="high"
        else
            profile="main"
        fi
        echo "====== Transcoding file ${codec[$k]}.mp4 to ${codec[$k]}_${size[$i]}.mp4 with Max Bandwidth ${bandwidth[$i]} ========="
        ffmpeg -i ${codec[$k]}.mp4 -c:v ${codec[$k]} -c:a copy -preset medium -b:v ${bandwidth[$i]} -bufsize ${buffer[$i]} \
            -vf scale=${width[$i]}:${height[$i]} -profile:v $profile -tune psnr \
            -vsync 0 -stats -y ${codec[$k]}_${size[$i]}.mp4
    done
done


for ((k=0;k<${#codec[@]};k++)); do
    # For different codec
    for ((j=0;j<${#size[@]};j++)); do
    # For different input file size
        for ((i=0;i<${#size[@]};i++)); do
            # For different recode parameters - width, height, bandwidth

            if [ "${codec[$k]}" = "libx264" ]; then
                profile="high"
                #threads=$(($vcpu*2))
                #buffer_latency=(${buffer_latency_264[*]})
            else
                profile="main"
               #threads=$(($vcpu))
               # buffer_latency=(${buffer_latency_265[*]})
            fi

            for ((m=0;m<${#use_profile[@]};m++)); do
                # For different user profile - high quality or low latency
                
                if [ "${use_profile[$m]}" = "quality" ]; then
                    ffmpeg_para="-preset medium -b:v ${bandwidth[$i]} -bufsize ${buffer[$i]} -tune psnr"
                else
                    ffmpeg_para="-preset fast -b:v ${bandwidth[$i]} -bufsize ${buffer[$i]} -g 999999 -x264opts no-sliced-threads:nal-hrd=cbr -tune psnr"
                fi

                echo "====== Transcoding For Benchmark Data File ${codec[$k]}_${size[$j]}.mp4 with Size ${size[$i]} Max Bandwidth ${bandwidth[$i]} ${use_profile[$m]} ========="
                # Recode with codec h264 or 265
                (time ffmpeg -i ${codec[$k]}_${size[$j]}.mp4 -c:v ${codec[$k]} -c:a copy -vf scale=${width[$i]}:${height[$i]} \
                    -profile:v $profile $ffmpeg_para \
                    -vsync 0 -v quiet -stats -y $path/${codec[$k]}_${size[$j]}-${size[$i]}-${use_profile[$m]}.mp4 \
                    2>&1 ) > $path/${codec[$k]}_${size[$j]}-${size[$i]}-${use_profile[$m]}.log 2>&1

                # Get ffmpeg speed 
                speed=`cat $path/${codec[$k]}_${size[$j]}-${size[$i]}-${use_profile[$m]}.log | awk -F'=' '/speed/ {print $NF}'`
                # Get CPU usage
                user=`cat $path/${codec[$k]}_${size[$j]}-${size[$i]}-${use_profile[$m]}.log | awk '/user/ {print $2}' | awk -F'm' '{printf 60*int($1*100)/100+int($2*100)/100}'`
                real=`cat $path/${codec[$k]}_${size[$j]}-${size[$i]}-${use_profile[$m]}.log | awk '/real/ {print $2}' | awk -F'm' '{printf 60*int($1*100)/100+int($2*100)/100}'`
                cpu_usage=`echo "scale=4;$user/$real/$vcpu/2*100;" | bc`

                # Log cpu_usage and speed
                echo "${codec[$k]}_${size[$j]}-${size[$i]} ${use_profile[$m]} cpu_usage $cpu_usage speed $speed" >> ${codec[$k]}_benchmark.log
                cat ${codec[$k]}_benchmark.log

            done
        done
    done
done

