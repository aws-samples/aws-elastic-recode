// Copyright 2020 AWS ElasticRecode Solution Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// @author Su Wei <suwei007@gmail.com>

package model

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/aws/aws-sdk-go/aws"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

//EC2Profile is model struct
type EC2Profile struct {
	PriceModel   *string `json:"priceModel,omitempty"`
	InstanceType *string `json:"instanceType,omitempty"`
}

//FFmpegProfile  is model struct
type FFmpegProfile struct {
	Codec       *string `json:"codec,omitempty"`
	OriginCodec *string `json:"originCodec,omitempty"`
	Scale       *string `json:"scale,omitempty"`
	Bitrate     *string `json:"bitrate,omitempty"`
	BufferSize  *string `json:"buffersize,omitempty"`
	Profile     *string `json:"profile,omitempty"`
	Platform    *string `json:"platform,omitempty"`
}

//JobProfile is API
type JobProfile struct {
	EC2Profile    *EC2Profile    `json:"ec2,omitempty"`
	FFmpegProfile *FFmpegProfile `json:"ffmpeg,omitempty"`
	VMAFProfile   *VMAFProfile   `json:"vmaf,omitempty"`
}

//Job is model struct use UI and SQS_Worker
type Job struct {
	UserID      *string     `json:"userid,omitempty"`
	JobID       *string     `json:"jobid,omitempty"`
	Input       *string     `json:"input,omitempty"`
	BatchInputs []string    `json:"batchInputs,omitempty"`
	Output      *string     `json:"output,omitempty"`
	Priority    int         `json:"priority,omitempty"`
	Profile     *JobProfile `json:"profile,omitempty"`
}

type JobOutput struct {
	UserID *string    `json:"userid,omitempty"`
	JobID  *string    `json:"jobid,omitempty"`
	Input  *string    `json:"input,omitempty"`
	Output *string    `json:"output,omitempty"`
	Status *JobStatus `json:"status,omitempty"`
}

type JobStatus struct {
	Event   *string `json:"event,omitempty"`
	Status  *string `json:"status,omitempty"`
	Message *string `json:"message,omitempty"`
}

type VMAFJob struct {
	UserID     *string     `json:"userid,omitempty"`
	JobID      *string     `json:"jobid,omitempty"`
	InputFile  *string     `json:"input,omitempty"`
	OriginFile *string     `json:"origin,omitempty"`
	Output     *string     `json:"output,omitempty"`
	Profile    *JobProfile `json:"profile,omitempty"`
	Status     *JobStatus  `json:"status,omitempty"`
}

type VMAFProfile struct {
	Scale  *string `json:"scale,omitempty"`
	SSIM   *string `json:"ssim,omitempty"`
	PSNR   *string `json:"psnr,omitempty"`
	MSSSIM *string `json:"ms-ssim,omitempty"`
}

const JobEventSummit = "JobSummit"
const JobEventStart = "JobStart"
const JobEventFinished = "JobFinished"

func int2size(size int) string {
	if size < 1024 {
		return fmt.Sprintf("%dK", size)
	}
	return fmt.Sprintf("%dM", size/1024)
}

func size2int(size string) (int, error) {

	invalidByteQuantityError := fmt.Errorf("invalid format %s,  etc. 1K,1KB, 1M, 1MB", size)
	size = strings.TrimSpace(size)
	size = strings.ToUpper(size)

	i := strings.IndexFunc(size, unicode.IsLetter)

	if i == -1 {
		return 0, invalidByteQuantityError
	}

	bytesString, multiple := size[:i], size[i:]
	bytes, err := strconv.ParseFloat(bytesString, 64)
	if err != nil || bytes < 0 {
		return 0, invalidByteQuantityError
	}

	switch multiple {
	case "M", "MB", "MIB":
		return int(bytes * 1024), nil
	case "K", "KB", "KIB":
		return int(bytes), nil
	default:
		return 0, invalidByteQuantityError
	}

}

func (j *VMAFJob) prepareCommonProfile() error {
	if j.Profile == nil || j.Profile.VMAFProfile == nil {
		return fmt.Errorf("not found profile %s, %s", *j.UserID, *j.JobID)
	}
	if _, exits := vmafScaleMap[*j.Profile.VMAFProfile.Scale]; !exits {
		return fmt.Errorf("scale config error %s, %s, %s", *j.UserID, *j.JobID, *j.Profile.VMAFProfile.Scale)
	}
	j.Profile.VMAFProfile.Scale = aws.String(vmafScaleMap[*j.Profile.VMAFProfile.Scale])

	return nil
}

func (j *Job) prepareCommonProfile() error {

	var scale string
	var exits bool

	platform := *j.Profile.FFmpegProfile.Platform
	profile := *j.Profile.FFmpegProfile.Profile

	//check platform ["CPU", "GPU"]
	if !HasElem(ec2Platform, platform) {
		return fmt.Errorf("Profile build Wrong, invalid FFmpeg platform: %s ", *j.Profile.FFmpegProfile.Platform)
	}

	//check profile ["quality", "latency"]
	if !HasElem(ffmpegProfile, *j.Profile.FFmpegProfile.Profile) {
		return fmt.Errorf("Profile build Wrong, invalid FFmpeg profile: %s ", *j.Profile.FFmpegProfile.Profile)
	}

	bitrate, err := size2int(*j.Profile.FFmpegProfile.Bitrate)
	if err != nil {
		return fmt.Errorf("Profile build Wrong, invalid bitrate : %s ", *j.Profile.FFmpegProfile.BufferSize)
	}

	//check job priority [0,100]
	if j.Priority < 0 || j.Priority > 100 {
		return fmt.Errorf("Profile build Wrong, invalid Job Priority: %d ", j.Priority)
	}

	//check scale ["1080p", "720p", "480p", "360p", "240p"]
	if scale, exits = ffmpegScaleMap[*j.Profile.FFmpegProfile.Scale]; !exits {
		return fmt.Errorf("Profile build Wrong, invalid Scale: %s ", *j.Profile.FFmpegProfile.Scale)
	}

	//check profile , set buffersize
	if profile == "quality" {
		j.Profile.FFmpegProfile.BufferSize = aws.String(int2size(bitrate * 2))
	}

	if profile == "latency" {
		j.Profile.FFmpegProfile.BufferSize = aws.String(int2size(bitrate / 25))
	}

	j.Profile.FFmpegProfile.Scale = aws.String(scale)

	if platform == "cpu" {
		return j.prepareCPUProfile()
	}

	if platform == "gpu" {
		if j.Profile.FFmpegProfile.OriginCodec == nil {
			j.Profile.FFmpegProfile.OriginCodec = aws.String("h264")
		}
		return j.prepareGPUProfile()
	}

	return nil

}

func (j *Job) prepareCPUProfile() error {

	var codec string
	var exits bool

	if codec, exits = ffmpegCPUCodesMap[*j.Profile.FFmpegProfile.Codec]; !exits {
		return fmt.Errorf("CPUProfile build Wrong, invalid CODEC: %s ", *j.Profile.FFmpegProfile.Codec)
	}

	j.Profile.FFmpegProfile.Codec = aws.String(codec)

	return nil
}

func (j *Job) prepareGPUProfile() error {

	var codec string
	var exits bool

	if codec, exits = ffmpegGPUCodesMap[*j.Profile.FFmpegProfile.Codec]; !exits {
		return fmt.Errorf("GPUProfile build Wrong, invalid CODEC: %s ", *j.Profile.FFmpegProfile.Codec)
	}

	j.Profile.FFmpegProfile.Codec = aws.String(codec)

	//设置originCodec
	if codec, exits = ffmpegGPUOriginCodesMap[*j.Profile.FFmpegProfile.OriginCodec]; !exits {
		return fmt.Errorf("CPUProfile build Wrong, invalid CODEC: %s ", *j.Profile.FFmpegProfile.Codec)
	}

	j.Profile.FFmpegProfile.Codec = aws.String(codec)

	return nil
}

func HasElem(s interface{}, elem interface{}) bool {
	arrV := reflect.ValueOf(s)

	if arrV.Kind() == reflect.Slice {
		for i := 0; i < arrV.Len(); i++ {

			// XXX - panics if slice element points to an unexported struct field
			// see https://golang.org/pkg/reflect/#Value.Interface
			if arrV.Index(i).Interface() == elem {
				return true
			}
		}
	}

	return false
}

//NewJob build job from job
func NewJob(data []byte) (*Job, error) {

	job := &Job{}
	err := json.Unmarshal(data, job)
	if err != nil {
		return nil, err
	}

	u1 := uuid.Must(uuid.NewV4())
	job.JobID = aws.String(u1.String())
	err = job.prepareCommonProfile()
	if err != nil {
		return nil, err
	}

	return job, nil
}

//NewVMAFJob build job from job
func NewVMAFJob(data []byte) (*VMAFJob, error) {

	job := &VMAFJob{}
	err := json.Unmarshal(data, job)
	if err != nil {
		return nil, err
	}

	u1 := uuid.Must(uuid.NewV4())
	job.JobID = aws.String(u1.String())
	err = job.prepareCommonProfile()
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (job *Job) IsBatchJob() bool {
	if job.Input == nil && len(job.BatchInputs) > 0 {
		return true
	}
	return false
}

func (job *Job) BuildBatchJobs() ([]*Job, error) {
	if !job.IsBatchJob() {
		data, err := json.Marshal(job)
		if err != nil {
			log.Errorf("Job marshal failed [userid:%s] ", *job.UserID)
		}
		log.Errorf("Job [userid:%s] is not batch job ", *job.UserID)
		log.Debugf("Job %s", string(data))
		return nil, fmt.Errorf("Job [userid:%s] is not batch job ", *job.UserID)

	}

	batchInputs := job.BatchInputs
	jobs := []*Job{}
	for idx := range batchInputs {
		u1 := uuid.Must(uuid.NewV4())

		jobs = append(jobs, &Job{

			UserID:      aws.String(*job.UserID),
			JobID:       aws.String(u1.String()),
			Input:       aws.String(batchInputs[idx]),
			BatchInputs: nil,
			Output:      aws.String(*job.Output),
			Priority:    job.Priority,
			Profile:     job.Profile,
		})
	}

	return jobs, nil
}
