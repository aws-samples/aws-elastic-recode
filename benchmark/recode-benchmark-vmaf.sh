#!/bin/bash

# BEFORE running this script:
# 1. Put reference video in the same diretory with this script
# 2. Use the following filename for reference video
    ## ref.mp4
# 3. The script will generate the test video main.mp4, 
#    and then run vmaf scores with the test video and reference video
# 4. Use -1080 to run vmaf scores with 1080p reference video,
#    and -720 for 720p reference video, please ensure that the
#    reference video is 1080p or 720p. If not specified, will
#    use -1080 by default
# 5. Use -ssim to enable metric ssim, and -psnr for metric psnr,
#    which are optional

if [ -z $1 ]; then
    echo "Usage: Run benchmark.sh [-1080|-720] [-ssim] [-psnr]"
    exit -1
fi

ssim_enabled=0
psnr_enabled=0
ref_video="ref.mp4"
test_video="main.mp4"
bitrate="2.5M"
codec="libx264"
scale="1280:720"    # generate 720p test video to run vmaf with 1080p reference video
vmaf_param="[0:v]scale=1920x1080:flags=bicubic[main];[main][1:v]libvmaf=log_path=vmaf_scores.json"

for arg in $*
do
    if [ $arg = '-720' ];then
        vmaf_param="[0:v]scale=1280x720:flags=bicubic[main];[main][1:v]libvmaf=log_path=vmaf_scores.json"
        bitrate="1M" 
        scale="800:480"     # generate 480p test video to run vmaf with 720p reference video
    elif [ $arg = '-1080' ];then
        vmaf_param="[0:v]scale=1920x1080:flags=bicubic[main];[main][1:v]libvmaf=log_path=vmaf_scores.json"    
    elif [ $arg = '-ssim' ];then
        ssim_enabled=1        
    elif [ $arg = '-psnr' ];then
        psnr_enabled=1
    else
        echo "Invalid parameter $arg, Usage: Run benchmark.sh [-1080|-720] [-ssim] [-psnr]"
        exit -1
    fi
done

if [ $psnr_enabled = 1 ];then
    vmaf_param=${vmaf_param}":psnr=1"
fi

if [ $ssim_enabled = 1 ];then
    vmaf_param=${vmaf_param}":ssim=1"
fi

echo "====== Generating test video $test_video from reference video $ref_video ========="

ffmpeg -i $ref_video -c:v $codec -c:a copy -preset medium -b:v $bitrate -bufsize $bitrate \
    -vf scale=$scale -profile:v high -tune psnr \
    -vsync 0 -v quiet -stats -y $test_video

echo "====== Run VMAF scores for test video $test_video ========="

(time ffmpeg -i main.mp4 -i ref.mp4 -filter_complex $vmaf_param -f null - \
    2>&1 | grep speed ) > temp.log 2>&1

# get vmaf speed
speed=`cat temp.log | awk -F'=' '/speed/ {print $NF}'`

# log speed info
echo "speed:$speed" > benchmark_vmaf.log
cat benchmark_vmaf.log
