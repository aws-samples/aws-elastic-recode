package service

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	uuid "github.com/satori/go.uuid"
	"github.com/stevensu1977/elasticrecode/pkg/controlplane/config"
	"github.com/stevensu1977/elasticrecode/pkg/controlplane/model"
)

func ProcessJob(jobs []*model.Job) []*model.JobOutput {

	globalConfig := config.InitConfig()

	result := []*model.JobOutput{}

	for idx := range jobs {
		u1 := uuid.Must(uuid.NewV4())
		jobs[idx].JobID = aws.String(u1.String())
		status := &model.JobStatus{
			Event:  aws.String("JobSummit"),
			Status: aws.String("success"),
		}
		//根据Job的platform参数(CPU|GPU)随机挑选Job队列进行发送
		queue, err := globalConfig.PickWorkerQueue(strings.ToLower(*jobs[idx].Profile.FFmpegProfile.Platform))
		if err != nil {
			status.Status = aws.String("failed")
			status.Message = aws.String(err.Error())
		} else {
			err = InitSQSClient().SendMessage(queue, jobs[idx])
			if err != nil {
				status.Status = aws.String("failed")
				status.Message = aws.String(err.Error())
			}

		}

		jobOutput := &model.JobOutput{
			UserID: jobs[idx].UserID,
			JobID:  jobs[idx].JobID,
			Input:  jobs[idx].Input,
			Output: jobs[idx].Output,
			Status: status,
		}
		result = append(result, jobOutput)
	}

	return result
}

func ProcessVMAFJob(job *model.VMAFJob) *model.VMAFJob {

	globalConfig := config.InitConfig()

	status := &model.JobStatus{
		Event:  aws.String("JobSummit"),
		Status: aws.String("success"),
	}

	//根据Job的platform参数(CPU|GPU)随机挑选Job队列进行发送
	queue, err := globalConfig.PickVMAFQueue()
	if err != nil {
		status.Status = aws.String("failed")
		status.Message = aws.String(err.Error())
	} else {
		err = InitSQSClient().SendMessage(queue, job)
		if err != nil {
			status.Status = aws.String("failed")
			status.Message = aws.String(err.Error())
		}

	}

	job.Status = status

	return job
}
