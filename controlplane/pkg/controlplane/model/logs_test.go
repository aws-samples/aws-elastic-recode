package model

import (
	"encoding/json"
	"testing"
)

var jobLogs = `
{
	"workerType": "ffmpeg", 
	"nodeName": "ip-192-168-63-212.ec2.internal", 
	"podName": "ffmpeg-worker-95d8865bb-gtpnv", 
	"action": "JobStart", 
	"job": 
	{   "userid": "8242d788-3577-455b-9927-12fa48c52fe7", 
		"jobid": "d17b40f4-e7af-42aa-8ac0-22398e7b1b6b", 
		"input": "s3://m.azeroth.one/static/video/example.mp4", 
		"output": "s3://m.azeroth.one/static/video/output/", 
		"priority": 100, 
		"profile": {
			"ffmpeg": 
			{"codec": "libx264", 
			"scale": "1920:1080", 
		    "bitrate": "1M", "buffersize": "2M", 
	        "profile": "quality", "platform": 
            "cpu"
           }
       }
   }, 
"msg": "", 
"ts": 1588730504
}
`

var wrongJobLogs = `
{
	"workerType": "ffmpeg", 
	"nodeName": "ip-192-168-63-212.ec2.internal", 
	"podName": "ffmpeg-worker-95d8865bb-gtpnv", 
	"action": "JobStart", 
	 
"msg": "", 
"ts": 1588730504
}
`

func TestJobJSON(t *testing.T) {
	job := &JobLog{}
	err := json.Unmarshal([]byte(jobLogs), job)
	if err != nil {
		t.Fatal(err)
	}

	data, err := json.Marshal(job)

	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(data))

}

func TestNewJobLog(t *testing.T) {
	jobLog, err := NewJobLog([]byte(wrongJobLogs))

	if err != nil {
		t.Fatalf("%s", err)
	}
	t.Logf("%v", jobLog)
}
