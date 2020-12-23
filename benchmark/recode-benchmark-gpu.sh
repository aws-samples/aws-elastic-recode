#!/bin/bash

# BEFORE running this script:
# 1. Put origin videos in the same diretory with this script
# 2. Use the following filename for origin video
    ## h264.mp4 - for 264 video
    ## h265.mp4 - for 265 video


if [ -z $1 ]; then
    echo "Usage: Run ./bench.sh <result-path>"
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
buffer_quality=(8M 5M 1800K 1200K)
buffer_latency_264=(166K 104K 38K 25K)
buffer_latency_265=(2M 1250K 450K 300K)
codec=(h264 h265)
use_profile=(quality latency)

# Prepare for input files of different resolutions
for ((k=0;k<${#codec[@]};k++)); do
    # For different codec
    echo "====== Transcoding Input Files ========="
    for ((i=0;i<${#size[@]};i++)); do

        if [ "${codec[$k]}" = "h264" ]; then
            dec_codec="h264_cuvid"
            enc_codec="h264_nvenc"
            profile="high"
        else
            dec_codec="hevc_cuvid"
            enc_codec="hevc_nvenc"
            profile="main"
        fi    
        echo "====== Transcoding file ${codec[$k]}.mp4 to ${codec[$k]}_${size[$i]}.mp4 with Max Bandwidth ${bandwidth[$i]} ========="
        ffmpeg -hwaccel cuvid -c:v $dec_codec -i ${codec[$k]}.mp4 -c:v $enc_codec -c:a copy -preset medium -b:v ${bandwidth[$i]} -bufsize ${buffer_quality[$i]} \
            -filter:v scale_npp=w=${width[$i]}:h=${height[$i]} -profile:v $profile -bf 3 -b_ref_mode 2 -temporal-aq 1 -rc-lookahead 20 \
            -vsync 0 -stats -y ${codec[$k]}_${size[$i]}.mp4
    done
done



for ((k=0;k<${#codec[@]};k++)); do
    # For different codec
    for ((j=0;j<${#size[@]};j++)); do
    # For different input file size
        for ((i=0;i<${#size[@]};i++)); do
            # For different recode parameters - width, height, bandwidth

            if [ "${codec[$k]}" = "h264" ]; then
                dec_codec="h264_cuvid"
                enc_codec="h264_nvenc"
                profile="high"
                buffer_latency=(${buffer_latency_264[*]})
            else
                dec_codec="hevc_cuvid"
                enc_codec="hevc_nvenc"               
                profile="main"
                buffer_latency=(${buffer_latency_265[*]})
            fi

           for ((m=0;m<${#use_profile[@]};m++)); do
                # For different user profile - high quality or low latency
                
                if [ "${use_profile[$m]}" = "quality" ]; then
                    ffmpeg_para="-preset medium -b:v ${bandwidth[$i]} -bufsize ${buffer_quality[$i]} -bf 3 -b_ref_mode 2 -temporal-aq 1 -rc-lookahead 20"
                else
                    ffmpeg_para="-preset llhp -b:v ${bandwidth[$i]} -bufsize ${buffer_latency[$i]} -rc cbr_ld_hq -g 999999"
                fi

                echo "====== Transcoding For Benchmark Data File ${codec[$k]}_${size[$j]}.mp4 with Size ${size[$i]} Max Bandwidth ${bandwidth[$i]} ${use_profile[$m]} ========="
                # Recode with codec h264 or 265
                nohup ffmpeg -hwaccel cuvid -c:v $dec_codec -i ${codec[$k]}_${size[$j]}.mp4 -c:v $enc_codec -c:a copy -profile:v $profile \
                    -filter:v scale_npp=w=${width[$i]}:h=${height[$i]} $ffmpeg_para  \
                    -vsync 0 -stats -y $path/${codec[$k]}_${size[$j]}-${size[$i]}-${use_profile[$m]}.mp4 \
                    2>&1  > $path/${codec[$k]}_${size[$j]}-${size[$i]}-${use_profile[$m]}.log &

                # Get GPU utilization
                gpu_usage=`nvidia-smi dmon -c 2 | tail -n 1 | awk '{print "gpu_enc "$7" gpu_dec "$8}'`
            
            
                # Get ffmpeg speed after ffmpeg is done
                ps=`ps aux | grep -v grep | grep ffmpeg`
                
                while [[ -n "$ps" ]]
                do
                    sleep 10
                    ps=`ps aux | grep -v grep | grep ffmpeg`
                done
                
                speed=`cat $path/${codec[$k]}_${size[$j]}-${size[$i]}-${use_profile[$m]}.log | awk -F'=' '/speed/ {print $NF}'`

                # Log cpu_usage and speed
                echo "${codec[$k]}_${size[$j]}-${size[$i]} ${use_profile[$m]} $gpu_usage speed $speed" >> ${codec[$k]}_benchmark.log
                cat ${codec[$k]}_benchmark.log
            done
           
        done
    done
done

