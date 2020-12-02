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
	"encoding/json"
	"net/http"

	"github.com/stevensu1977/elasticrecode/pkg/controlplane/config"
	"github.com/stevensu1977/elasticrecode/pkg/controlplane/model"
)

//APIVersion func is wrapper func add Content-Type Application/JSON
func APIVersion(w http.ResponseWriter, r *http.Request) {

	data, _ := json.Marshal(map[string]string{
		"server":  model.ServerName,
		"version": model.ServerVersion,
	})
	write(w, data)

}

func GetWorkerQueues(w http.ResponseWriter, r *http.Request) {

	globalConfig := config.InitConfig()
	queues := make(map[string]interface{})

	queues["cpu"] = globalConfig.CPUQueues
	queues["gpu"] = globalConfig.GPUQueues
	queues["vmaf"] = globalConfig.VMAFQueues

	data, _ := json.Marshal(queues)
	write(w, data)

}

func GetWorkerDeployments(w http.ResponseWriter, r *http.Request) {

	globalConfig := config.InitConfig()

	data, _ := json.Marshal(globalConfig.WorkerDeployments)
	write(w, data)

}
