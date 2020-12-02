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
	"github.com/aws/aws-sdk-go/aws"
)

//WorkerNode is worker model
type WorkerNode struct {
	Name         *string
	Platform     *string
	InstanceType *string
	PriceModel   *string
}

//WorkerDeployment deployment
type WorkerDeployment struct {
	Name      *string
	WokerType *string
	Queue     *string
	Platform  *string
	Replicas  int
}

func NewWorkerDeployment(queueName, workerType, platform, deployment string, replicas int) *WorkerDeployment {

	return &WorkerDeployment{
		Name:      aws.String(deployment),
		WokerType: aws.String(workerType),
		Queue:     aws.String(queueName),
		Platform:  aws.String(platform),
		Replicas:  replicas,
	}
}
