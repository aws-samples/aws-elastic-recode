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
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/stevensu1977/elasticrecode/pkg/controlplane/config"
	"github.com/stevensu1977/elasticrecode/pkg/controlplane/service"
)

//GetLogs func is wrapper func add Content-Type Application/JSON
func GetLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	log.Debugf("GetLogs,  %v", vars)
	key := "userID"
	if _, ok := vars["jobID"]; ok {
		key = "jobID"
	}

	globalConfig := config.InitConfig()
	if !globalConfig.WriteJobLogs {
		writeError(w, http.StatusBadRequest, fmt.Errorf("disableWriteJobLogs is %v, logs api is not work", !globalConfig.WriteJobLogs))
		return
	}

	logs, err := service.FindJobLogs(key, vars[key])
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	data, _ := json.Marshal(logs)
	write(w, data)

}
