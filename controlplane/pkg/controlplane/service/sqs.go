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
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	log "github.com/sirupsen/logrus"
	"github.com/stevensu1977/elasticrecode/pkg/controlplane/config"
)

//SQSClient is helper struct
type SQSClient struct {
	svc    *sqs.SQS
	queues map[string]*string
}

const timeoutPtr = 20

var sqsClient *SQSClient

//InitSQSClient build single client
func InitSQSClient() *SQSClient {

	if sqsClient == nil {
		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))
		svc := sqs.New(sess)
		sqsClient = &SQSClient{
			svc:    svc,
			queues: map[string]*string{},
		}
		config := config.InitConfig()

		for idx := range config.CPUQueues {
			err := sqsClient.AddQueue(config.CPUQueues[idx])
			if err != nil {
				continue
			}
		}

		for idx := range config.GPUQueues {
			err := sqsClient.AddQueue(config.GPUQueues[idx])
			if err != nil {
				continue
			}
		}

	}
	return sqsClient

}

//SubQueue is queue consumer
func (c *SQSClient) SubQueue(queueName string, output chan string) error {

	_, err := c.svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		return err
	}

	go func() {
		err := c.receiveMessage(queueName, output)
		if err != nil {
			log.Errorf("AWS SQS: subscribe [%s] failed", err.Error())
			return
		}
	}()
	log.Infof("AWS SQS: subscribe [%s] successed", queueName)
	return nil
}

//GetQueue get queue url
func (c *SQSClient) getQueue(queueName string) (*string, bool) {
	v, found := c.queues[queueName]

	if found {
		return v, found
	}

	result, err := c.svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		return nil, false
	}
	c.queues[queueName] = result.QueueUrl
	return c.queues[queueName], true
}

//AddQueue check queue url
func (c *SQSClient) AddQueue(queueName string) error {
	if _, exits := c.queues[queueName]; exits {
		return nil
	}

	resultURL, err := c.svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		return err
	}
	log.Debugf("AddQueue %s,%s", queueName, *resultURL.QueueUrl)
	c.queues[queueName] = resultURL.QueueUrl
	return nil
}

//SendMessage is SQS send func
func (c *SQSClient) SendMessage(queueName string, data interface{}) error {
	queueURL, exits := c.getQueue(queueName)
	if !exits {
		return fmt.Errorf("Queue %s not exits", queueName)
	}
	input, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = c.svc.SendMessage(&sqs.SendMessageInput{
		DelaySeconds: aws.Int64(10),
		MessageBody:  aws.String(string(input)),
		QueueUrl:     queueURL,
	})
	return err
}

func (c *SQSClient) SendMessageManual(queueName string, data interface{}) error {

	resultURL, err := c.svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		return err
	}

	input, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = c.svc.SendMessage(&sqs.SendMessageInput{
		DelaySeconds: aws.Int64(10),
		MessageBody:  aws.String(string(input)),
		QueueUrl:     resultURL.QueueUrl,
	})
	return err
}

func (c *SQSClient) receiveMessage(queueName string, output chan string) error {
	config := config.InitConfig()

	queueURL, err := c.svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})

	if err != nil {
		return fmt.Errorf("Queue %s not exits", queueName)
	}

	for {
		result, err := c.svc.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl: queueURL.QueueUrl,
			AttributeNames: aws.StringSlice([]string{
				"SentTimestamp",
			}),
			MaxNumberOfMessages: aws.Int64(1),
			MessageAttributeNames: aws.StringSlice([]string{
				"All",
			}),
			WaitTimeSeconds: aws.Int64(timeoutPtr),
		})
		if err != nil {
			log.Errorf("Unable to receive message from queue %q, %v.", *queueURL, err)
		}

		if len(result.Messages) > 0 {
			log.Printf("Received %d messages.\n", len(result.Messages))
			log.Debug(result.Messages)
			for _, msg := range result.Messages {
				if config.WriteJobLogs {
					WriteJobLogs([]byte(*msg.Body))
				}
				output <- *msg.Body
				resultDelete, err := c.svc.DeleteMessage(&sqs.DeleteMessageInput{
					QueueUrl:      queueURL.QueueUrl,
					ReceiptHandle: msg.ReceiptHandle,
				})
				if err == nil {
					log.Println("Message Deleted", resultDelete)
				} else {
					log.Errorf("Message Deleted error %s", err.Error())
				}

			}
		}
	}

}
