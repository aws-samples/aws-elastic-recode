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

package service

import (
	"strings"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	log "github.com/sirupsen/logrus"

	"github.com/stevensu1977/elasticrecode/pkg/controlplane/config"
	"github.com/stevensu1977/elasticrecode/pkg/controlplane/model"
)

var clientset *kubernetes.Clientset

const prefix = "elastirecord.kubernetes.io"
const autodiscovery = prefix + "/" + "worker"
const annotationPlatform = prefix + "/" + "platform"

//InitK8SClientset K8S链接初始化
func InitK8SClientset() *kubernetes.Clientset {

	config := config.InitConfig()

	//AutoDiscovery 关闭模式
	if !config.AutoDiscovery {
		log.Infof("K8S: AutoDiscovery %v, , stop K8S Init", config.AutoDiscovery)
		return nil
	}

	if clientset == nil {
		var k8sConfig *rest.Config
		var err error
		if config.InCluster {
			k8sConfig, err = rest.InClusterConfig()
			if err != nil {
				log.Errorf("K8S: config init %s", err.Error())

			}

		} else {
			k8sConfig, err = clientcmd.BuildConfigFromFlags("", *config.KubeConfig)
			if err != nil {
				log.Errorf("K8S: config init %s", err.Error())

			}
		}

		clientset, err = kubernetes.NewForConfig(k8sConfig)
		if err != nil {
			log.Errorf("K8S: Config init %s", err.Error())
		}
	}

	go initEventListen(clientset)

	return clientset
}

//initEventListen 通过Deployment, SharedInformer 获取K8S集群事件
func initEventListen(clientset *kubernetes.Clientset) {

	log.Info("K8S: AutoDicovery is true, kubernetes eventhook start")
	factory := informers.NewSharedInformerFactory(clientset, 0)
	informer := factory.Apps().V1().Deployments().Informer()
	stopper := make(chan struct{})
	defer close(stopper)
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			d := obj.(*v1.Deployment)
			processDeployment("AddEvent", d)
		},
		UpdateFunc: func(old, new interface{}) {
			d := new.(*v1.Deployment)
			processDeployment("UpdateEvent", d)
		},
	})
	informer.Run(stopper)

}

//processDeployment autodiscovery 模式下对deployment事件进行处理
func processDeployment(event string, d *v1.Deployment) {
	config := config.InitConfig()
	sqsClient := InitSQSClient()

	annotations := d.GetAnnotations()
	if annotation, ok := annotations[autodiscovery]; ok {
		log.Printf("ElasticRecode deployment with '%s' annotation created in namespace %s, name %s.\n", autodiscovery, d.GetNamespace(), d.GetName())
		if strings.ToLower(annotation) == "enable" {
			queueName := ""
			workerType := ""
			isSameControlPlane := false
			platform := "cpu"

			if p, ok := annotations[annotationPlatform]; ok {
				platform = p
			}
			for _, env := range d.Spec.Template.Spec.Containers[0].Env {
				log.Debug(env.Name, env.Value)
				if env.Name == "QUEUE_NAME" {
					queueName = env.Value
				}
				if env.Name == "WORKER_TYPE" {
					workerType = env.Value
				}
				if env.Name == "CONTROL_PLANE" && env.Value == *config.Queue {
					isSameControlPlane = true
				}
			}

			if isSameControlPlane {

				err := sqsClient.AddQueue(queueName)
				if err != nil {
					log.Errorf("AutoDiscovery error: %s", err.Error())
				} else {
					worker := model.NewWorkerDeployment(queueName, workerType, platform, d.Name, int(*d.Spec.Replicas))
					config.AddWorkerDeployment(worker)
					log.Infof("K8S|%s: [%s] ,wrokertype %s , platform %s, queue %s , replicas %d", event, d.Name, workerType, platform, queueName, int(*d.Spec.Replicas))
				}
			}

		}

	}
}
