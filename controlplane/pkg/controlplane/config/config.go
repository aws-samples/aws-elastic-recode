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

package config

import (
	"flag"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"

	"k8s.io/client-go/util/homedir"

	"github.com/stevensu1977/elasticrecode/pkg/controlplane/model"
	"github.com/stevensu1977/elasticrecode/pkg/controlplane/utils"
)

//ControlPlaneConfig 配置模型
type ControlPlaneConfig struct {
	Addr              *string
	Queue             *string
	InCluster         bool
	KubeConfig        *string
	AutoDiscovery     bool
	CPUQueues         []string
	VMAFQueues        []string
	GPUQueues         []string
	WriteJobLogs      bool
	LogsTableName     *string
	Verbose           bool
	CORS              bool
	WorkerDeployments map[string]*model.WorkerDeployment
}

//globalConfig 全局配置对象
var globalConfig *ControlPlaneConfig

//InitConfig 初始化全局配置对象
func InitConfig() *ControlPlaneConfig {

	rand.Seed(time.Now().Unix())

	if globalConfig == nil {
		var kubeconfig *string
		//配置参数
		queueName := flag.String("q", "control_plane", "Queue name")
		addr := flag.String("addr", ":8080", "http service address")
		disableInCluster := flag.Bool("disableInCluster", false, "disable in cluster model, use kubeconfig access k8s")
		disableAutoDiscovery := flag.Bool("disableAutoDiscovery", false, "disable auto discovery,use manual setup")
		disableWriteJobLogs := flag.Bool("disableWriteJobLogs", false, "disable write job logs to dynamedb")
		disableCORS := flag.Bool("disableCORS", false, "disable CORS request")
		logsTableName := flag.String("LogsTableName", "elasticrecode_logs", "logs dynamedb table name ")
		cpuQueues := flag.String("cpuQ", "", "manual setup cpu worker queues")
		vmafQueues := flag.String("vmafQ", "", "manual setup VMAF job queues")
		gpuQueues := flag.String("gpuQ", "", "manual setup gpu worker queues")
		verbose := flag.Bool("verbose", false, "verbose model")

		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}

		flag.Parse()

		if *verbose {
			log.SetLevel(log.TraceLevel)
		}

		//如果禁用InCluster,则需要使用kubeconfig
		if *disableInCluster {
			if !*disableAutoDiscovery && *kubeconfig == "" {
				flag.PrintDefaults()
				utils.ExitErrorf("If --disableAutoDiscovery is false , kubeconfig should be needed")
			}
		}
		if *disableAutoDiscovery {
			if *cpuQueues == "" && *gpuQueues == "" {
				flag.PrintDefaults()
				utils.ExitErrorf("If --disableAutoDiscovery is true , must use --cpuQ or --gpuQ setup worker queue")
			}
		}

		cpuQueuesArray := []string{}
		qpuQueuesArray := []string{}
		vmafQueuesArray := []string{}

		if *cpuQueues != "" {
			cpuQueuesArray = strings.Split(*cpuQueues, ",")
		}

		if *gpuQueues != "" {
			qpuQueuesArray = strings.Split(*gpuQueues, ",")
		}

		if *vmafQueues != "" {
			vmafQueuesArray = strings.Split(*vmafQueues, ",")
		}

		globalConfig = &ControlPlaneConfig{
			Addr:          addr,
			Queue:         queueName,
			InCluster:     !*disableInCluster,
			KubeConfig:    kubeconfig,
			AutoDiscovery: !*disableAutoDiscovery,
			CPUQueues:     cpuQueuesArray,
			VMAFQueues:    vmafQueuesArray,
			GPUQueues:     qpuQueuesArray,

			WriteJobLogs:      !*disableWriteJobLogs,
			LogsTableName:     logsTableName,
			Verbose:           *verbose,
			CORS:              !*disableCORS,
			WorkerDeployments: make(map[string]*model.WorkerDeployment),
		}

		log.Printf("AWS %s %s starting...", model.ServerName, model.ServerVersion)
		log.Println("################################################")
		log.Printf("Addr: %s", *globalConfig.Addr)
		log.Printf("Queue: %s", *globalConfig.Queue)
		log.Printf("InCluster: %v", globalConfig.InCluster)
		if !globalConfig.InCluster {
			log.Printf("KubeConfig: %s", *globalConfig.KubeConfig)
		}

		log.Printf("AutoDiscovery: %v", globalConfig.AutoDiscovery)
		if !globalConfig.AutoDiscovery {
			log.Printf("CPUQueues: %v", globalConfig.CPUQueues)
			log.Printf("GPUQueues: %v", globalConfig.GPUQueues)
		}

		log.Printf("WriteJobLogs: %v", globalConfig.WriteJobLogs)
		if globalConfig.WriteJobLogs {
			log.Printf("LogsTableName: %s", *globalConfig.LogsTableName)
		}

		log.Println("################################################")
	}

	return globalConfig

}

func (c *ControlPlaneConfig) PickWorkerQueue(platform string) (string, error) {

	if platform == "cpu" {
		if len(c.CPUQueues) == 0 {
			return "", fmt.Errorf("not have any CPU worker queue")
		}
		return c.CPUQueues[rand.Intn(len(c.CPUQueues))], nil
	}

	if platform == "gpu" {
		if len(c.GPUQueues) == 0 {
			return "", fmt.Errorf("not have any GPU worker queue")
		}
		return c.GPUQueues[rand.Intn(len(c.GPUQueues))], nil
	}

	return "", fmt.Errorf("not support platform %s", platform)

}

func (c *ControlPlaneConfig) PickVMAFQueue() (string, error) {

	if len(c.VMAFQueues) == 0 {
		return "", fmt.Errorf("not have any VMAF job queue")
	}
	return c.VMAFQueues[rand.Intn(len(c.VMAFQueues))], nil
}

func (c *ControlPlaneConfig) AddWorkerDeployment(workerDeployment *model.WorkerDeployment) {
	var exits bool
	_, exits = c.WorkerDeployments[*workerDeployment.Name]

	if exits {
		return
	}

	//添加worker队列
	if *workerDeployment.Platform == "cpu" {
		log.Infof("Add CPUQueue %s ", *workerDeployment.Queue)
		if *workerDeployment.WokerType == "ffmpeg" {
			if !HasElem(c.CPUQueues, *workerDeployment.Queue) {
				c.CPUQueues = append(c.CPUQueues, *workerDeployment.Queue)
			}
		}

		if *workerDeployment.WokerType == "vmaf" {
			if !HasElem(c.VMAFQueues, *workerDeployment.Queue) {
				c.VMAFQueues = append(c.VMAFQueues, *workerDeployment.Queue)
			}
		}

	}

	if *workerDeployment.Platform == "gpu" {
		log.Infof("Add GPUQueue %s ", *workerDeployment.Queue)
		if !HasElem(c.GPUQueues, *workerDeployment.Queue) {
			c.GPUQueues = append(c.GPUQueues, *workerDeployment.Queue)
			log.Infof("Add CPUQueue %s", *workerDeployment.Queue)
		}
	}
	c.WorkerDeployments[*workerDeployment.Name] = workerDeployment

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
