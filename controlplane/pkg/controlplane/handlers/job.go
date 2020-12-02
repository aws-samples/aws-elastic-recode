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

package handlers

import (
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/stevensu1977/elasticrecode/pkg/controlplane/model"
	controlplane "github.com/stevensu1977/elasticrecode/pkg/controlplane/service"
)

//CreateJob func is wrapper func add Content-Type Application/JSON
func CreateJob(w http.ResponseWriter, r *http.Request) {
	var job *model.Job

	defer r.Body.Close()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	job, err = model.NewJob(data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	jobs := []*model.Job{job}

	if job.IsBatchJob() {
		log.Debugf("this is batch job, %v", job.BatchInputs)
		jobs, err = job.BuildBatchJobs()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	output := controlplane.ProcessJob(jobs)
	//返回数组
	serverJSON(w, r, output)
}

//CreateVMAFJob 创建VMAF的API
func CreateVMAFJob(w http.ResponseWriter, r *http.Request) {

	var vmafJob *model.VMAFJob

	defer r.Body.Close()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	vmafJob, err = model.NewVMAFJob(data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	output := controlplane.ProcessVMAFJob(vmafJob)

	//返回json
	serverJSON(w, r, output)
}
