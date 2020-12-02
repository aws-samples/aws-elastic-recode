package model

import (
	"encoding/json"
	"testing"
)

var jobInput = `{
	"userid":"8242d788-3577-455b-9927-12fa48c52fe7",
	"input":"s3://m.azeroth.one/static/video/example.mp4",
	"output":"s3://m.azeroth.one/static/video/output.mp4",
	"priority":100,
	"profile":{
		"ec2":{
			"priceModel":"spot"
		},
		"ffmpeg":{
			"codec":"h264",
			"scale":"1080p",
			"bitrate":"1M",
			"profile":"latency",
			"platform":"cpu"
		}
	}
}`

var jobBatchInput = `{
	"userid":"8242d788-3577-455b-9927-12fa48c52fe7",
	"batchInputs":[
		"s3://m.azeroth.one/static/video/example.mp4",
		"s3://m.azeroth.one/static/video/example1.mp4",
		"s3://m.azeroth.one/static/video/example2.mp4"
		],
	"output":"s3://m.azeroth.one/static/video/output.mp4",
	"priority":100,
	"profile":{
		"ec2":{
			"priceModel":"spot"
		},
		"ffmpeg":{
			"codec":"h264",
			"scale":"1080p",
			"bitrate":"1M",
			"profile":"latency",
			"platform":"cpu"
		}
	}
}`

var vamfJob = `{
    "userid":"8242d788-3577-455b-9927-12fa48c52fe7",
    "test_file":"s3://m.azeroth.one/static/video/example.mp4",
    "ref_file":"s3://m.azeroth.one/static/video/example.mp4",
    "output":"s3://m.azeroth.one/static/video/vmaf/",
    "profile":{
    "vmaf":{
       "scale":"1280x720",
       "ssim":"enable",
       "psnr":"enable",
       "ms-ssim":"enable"
     }
    }
    
}
`

func TestJobProfile(t *testing.T) {

	job, err := NewJob([]byte(jobInput))
	if err != nil {
		t.Fatal(err)
	}

	data, err := json.Marshal(job)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))
}

func TestVMAFJobProfile(t *testing.T) {

	vmaf := &VMAFJob{}
	json.Unmarshal([]byte(vamfJob), vmaf)
	data, err := json.Marshal(vmaf)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))
}

func TestJobBatchProcess(t *testing.T) {
	job, err := NewJob([]byte(jobBatchInput))
	if err != nil {
		t.Fatal(err)
	}

	t.Log(job.IsBatchJob())
	jobs, err := job.BuildBatchJobs()
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(jobs)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))
}

func TestSize2Int(t *testing.T) {

	sizes := []string{"1K", "1M", "5M", "10M"}
	for _, s := range sizes {
		size, err := size2int(s)
		if err != nil {
			t.Fatal(err)
		}

		t.Log(s, "|", size, int2size(size), int2size(size*2), int2size(size/25))

	}

}
