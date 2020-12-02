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
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/stevensu1977/elasticrecode/pkg/controlplane/config"
	"github.com/stevensu1977/elasticrecode/pkg/controlplane/model"
)

var userIDIndex = "userIDIndex"
var jobIDIndex = "jobIDIndex"

const MaxItem = 50

type Duration int
type Direction int

const (
	Hours Duration = iota
	Days
	Weeks
)

const (
	Before Direction = iota
	After
)

func (d Duration) duration() time.Duration {
	return [...]time.Duration{time.Hour, 24 * time.Hour, 24 * 7 * time.Hour}[d]
}

type DurationInput struct {
	Duration  Duration
	Direction Direction
	Count     int64
}

//SQSClient is helper struct
type DynamoDBClient struct {
	svc *dynamodb.DynamoDB
}

var dynamoDBClient *DynamoDBClient

func (d *DurationInput) GetDuration() time.Duration {
	if d.Direction == Before {
		return (-1) * time.Duration(d.Count) * d.Duration.duration()
	}

	return time.Duration(d.Count) * d.Duration.duration()
}

func (d *DurationInput) Unix() int64 {
	now := time.Now()
	return now.Add(d.GetDuration()).Unix()
}

//InitSQSClient build single client
func InitDynamoDBClient() *DynamoDBClient {

	if dynamoDBClient == nil {
		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))
		svc := dynamodb.New(sess)
		dynamoDBClient = &DynamoDBClient{
			svc: svc,
		}
	}
	return dynamoDBClient

}

func buildLogQuery(key, value string, duration DurationInput) *dynamodb.QueryInput {
	config := config.InitConfig()
	indexName := userIDIndex
	if key == "jobID" {
		indexName = jobIDIndex
	}

	return &dynamodb.QueryInput{
		TableName: config.LogsTableName,
		IndexName: aws.String(indexName),
		KeyConditions: map[string]*dynamodb.Condition{
			key: {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(value),
					},
				},
			},
			"ts": {
				ComparisonOperator: aws.String("GE"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						N: aws.String(fmt.Sprintf("%d", duration.Unix())),
					},
				},
			},
		},
		Limit: aws.Int64(MaxItem),
	}

}

func WriteJobLogs(data []byte) {
	job, err := model.NewJobLog([]byte(data))
	if err != nil {
		log.Errorf("Got error calling PutItem: %s", err.Error())
		return
	}

	client := InitDynamoDBClient()
	config := config.InitConfig()

	item, err := dynamodbattribute.MarshalMap(job)
	if err != nil {
		log.Errorf("Got error calling PutItem: %s", err.Error())
	}

	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: config.LogsTableName,
	}

	_, err = client.svc.PutItem(input)
	if err != nil {
		log.Errorf("Got error calling PutItem: %s", err.Error())
	}
	log.Debugf("InsertLog %s successful", *job.JobID)
}

func FindJobLogs(key, value string) ([]model.JobLog, error) {
	client := InitDynamoDBClient()
	//默认显示1天的数据
	d := DurationInput{
		Count:     1,
		Duration:  Days,
		Direction: Before,
	}
	queryInput := buildLogQuery(key, value, d)

	resp1, err := client.svc.Query(queryInput)
	if err != nil {
		return nil, err
	} else {
		logs := []model.JobLog{}
		err = dynamodbattribute.UnmarshalListOfMaps(resp1.Items, &logs)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		_, err = json.Marshal(logs)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		return logs, nil
	}

}
