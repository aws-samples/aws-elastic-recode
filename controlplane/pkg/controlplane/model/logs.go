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
)

type JobLog struct {
	Timestamp *int64  `json:"ts,omitempty"`
	UserID    *string `json:"userID,omitempty"`
	JobID     *string `json:"jobID,omitempty"`
	Action    *string `json:"action,omitempty"`
	Job       *Job    `json:"job,omitempty"`
	NodeName  *string `json:"nodeName,omitempty"`
	PodName   *string `json:"podName,omitempty"`
}

type SysLog struct {
	Timestamp *int64  `json:"ts,omitempty"`
	Action    *string `json:"action,omitempty"`
	Raw       *string `json:"raw,omitempty"`
}

func NewJobLog(rawData []byte) (*JobLog, error) {
	job := &JobLog{}
	err := json.Unmarshal([]byte(rawData), job)
	if err != nil {
		return nil, err
	}
	//fix wrong job
	if job.Job == nil || job.Job.JobID == nil || job.Job.UserID == nil {
		wrontRawData, _ := json.Marshal(job)
		return nil, fmt.Errorf("this is wrong jobLog %s", wrontRawData)
	}
	if job.Job != nil {

		if job.Job.UserID != nil {
			job.UserID = job.Job.UserID
		}
		if job.Job.JobID != nil {
			job.JobID = job.Job.JobID
		}

	}
	return job, nil
}
