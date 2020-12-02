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

package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/stevensu1977/elasticrecode/pkg/controlplane/config"
	"github.com/stevensu1977/elasticrecode/pkg/controlplane/model"
	"github.com/stevensu1977/elasticrecode/pkg/controlplane/router"
	"github.com/stevensu1977/elasticrecode/pkg/controlplane/service"
)

var (
	buildTime  string
	commitHash string
)

func main() {

	fmt.Printf("version: %s\n", model.ServerVersion)
	fmt.Printf("build: %s\n", buildTime)
	fmt.Printf("git commit: %s\n", commitHash)

	//初始化所有参数
	globalConfig := config.InitConfig()
	log.Printf("AWS ElasticRecode ControlPlane listening on %s, EventQueue[%s]", *globalConfig.Addr, *globalConfig.Queue)

	//初始化K8S服务
	service.InitK8SClientset()

	//初始化SQS服务,订阅event queue
	output := make(chan string)
	go func() {
		err := service.InitSQSClient().SubQueue(*globalConfig.Queue, output)
		if err != nil {
			log.Panic(err)
		}
	}()

	//初始化web模块
	router.InitRoute(*globalConfig.Addr, output)

}
