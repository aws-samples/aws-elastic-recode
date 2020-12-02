package service

import (
	"testing"

	"github.com/stevensu1977/elasticrecode/pkg/controlplane/model"
)

var jobJSON = `{
	"userid": "8242d788-3577-455b-9927-12fa48c52fe7",
	"jobid": "12e7d74d-103b-4d96-84f3-324e899f9a59",
	"input": "s3://m.azeroth.one/static/video/example.mp4",
	"output": "s3://m.azeroth.one/static/video/output.mp4",
	"priority": 100,
	"profile": {
	  "ec2": {
		"priceModel": "spot"
	  },
	  "ffmpeg": {
		"codec": "libx264",
		"scale": "1920:1080",
		"bitrate": "1M",
		"bufsize": "2M",
		"profile": "quality",
		"platform": "cpu"
	  }
	}
  }`

func TestSendMessage(t *testing.T) {

	job, err := model.NewJob([]byte(jobJSON))
	if err != nil {
		t.Fatal(err)
	}

	client := InitSQSClient()
	err = client.SendMessage("cp_01", job)

	if err != nil {
		t.Fatal(err)
	}
}
