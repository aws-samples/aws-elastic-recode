package service

import (
	"encoding/json"

	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	log "github.com/sirupsen/logrus"

	"github.com/stevensu1977/elasticrecode/pkg/controlplane/model"
)

type Item struct {
	Timestamp int64 `json:"ts,omitempty"`
	UserID    string
	JobID     string
	Status    string
}

var tableName = "elasticrecode_logs"

var jobLogs = `
{
	"workerType": "ffmpeg", 
	"nodeName": "ip-192-168-63-212.ec2.internal", 
	"podName": "ffmpeg-worker-95d8865bb-gtpnv", 
	"action": "JobStart", 
	"job": 
	{   "userid": "8242d788-3577-455b-9927-12fa48c52fe7", 
		"jobid": "d17b40f4-e7af-42aa-8ac0-22398e7b1b6b", 
		"input": "s3://m.azeroth.one/static/video/example.mp4", 
		"output": "s3://m.azeroth.one/static/video/output/", 
		"priority": 100, 
		"profile": {
			"ffmpeg": 
			{"codec": "libx264", 
			"scale": "1920:1080", 
		    "bitrate": "1M", "buffersize": "2M", 
	        "profile": "quality", "platform": 
            "cpu"
           }
       }
   }, 
"msg": "", 
"ts": 1588730504
}
`

func TestDynamoDBClient(t *testing.T) {
	client := InitDynamoDBClient()
	t.Log("test dynamoDBClient", client)
}

func TestInsertItem(t *testing.T) {

	item := &Item{
		Timestamp: time.Now().Unix(),
		UserID:    "8242d788-3577-455b-9927-12fa48c52fe7",
		JobID:     "d8356414-53e4-4357-a1ba-46d6be3f18b5",
		Status:    "JobFinished",
	}

	client := InitDynamoDBClient()

	event, err := dynamodbattribute.MarshalMap(item)

	if err != nil {
		t.Fatal("Got error calling PutItem:", err.Error())
	}
	t.Log(tableName, "PutItem starting....")
	input := &dynamodb.PutItemInput{
		Item:      event,
		TableName: aws.String(tableName),
	}

	_, err = client.svc.PutItem(input)
	if err != nil {
		t.Fatal("Got error calling PutItem:", err.Error())
	}

}

func TestQueryItem(t *testing.T) {
	client := InitDynamoDBClient()

	userIDIndex := "UserID-Timestamp-index"
	userID := "8242d788-3577-455b-9927-12fa48c52fe7"

	var queryInput = &dynamodb.QueryInput{
		TableName: aws.String(tableName),
		IndexName: aws.String(userIDIndex),
		Limit:     aws.Int64(1),
		KeyConditions: map[string]*dynamodb.Condition{
			"UserID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(userID),
					},
				},
			},
		},
	}
	var resp1, err = client.svc.Query(queryInput)

	t.Log(resp1.LastEvaluatedKey)
	if err != nil {
		fmt.Println(err)
	} else {
		personObj := []Item{}
		err = dynamodbattribute.UnmarshalListOfMaps(resp1.Items, &personObj)
		if err != nil {
			t.Fatal(err)
		}
		log.Println(personObj)
	}

	if len(resp1.LastEvaluatedKey) > 0 {
		t.Log("have other page ")

		queryInput = &dynamodb.QueryInput{
			TableName:         aws.String(tableName),
			IndexName:         aws.String(userIDIndex),
			Limit:             aws.Int64(1),
			ExclusiveStartKey: resp1.LastEvaluatedKey,
			KeyConditions: map[string]*dynamodb.Condition{
				"UserID": {
					ComparisonOperator: aws.String("EQ"),
					AttributeValueList: []*dynamodb.AttributeValue{
						{
							S: aws.String(userID),
						},
					},
				},
			},
		}

		resp1, err = client.svc.Query(queryInput)
		if err != nil {
			t.Fatal(err)
		} else {
			personObj := []Item{}
			err = dynamodbattribute.UnmarshalListOfMaps(resp1.Items, &personObj)
			if err != nil {
				t.Fatal(err)
			}
			log.Println(personObj)
		}
	}

}

func TestJobLogInsert(t *testing.T) {
	job, err := model.NewJobLog([]byte(jobLogs))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(job)

	client := InitDynamoDBClient()

	event, err := dynamodbattribute.MarshalMap(job)

	t.Logf("%v", event)

	if err != nil {
		t.Fatal("Got error calling PutItem:", err.Error())
	}
	t.Log(tableName, "PutItem starting....")
	input := &dynamodb.PutItemInput{
		Item:      event,
		TableName: aws.String(tableName),
	}

	_, err = client.svc.PutItem(input)
	if err != nil {
		t.Fatal("Got error calling PutItem:", err.Error())
	}
}

func TestFindJobLogs(t *testing.T) {
	key := "userID"
	value := "8242d788-3577-455b-9927-12fa48c52fe7"

	client := InitDynamoDBClient()

	d := DurationInput{
		Count:     48,
		Duration:  Hours,
		Direction: Before,
	}

	queryInput := buildLogQuery(key, value, d)

	resp1, err := client.svc.Query(queryInput)
	if err != nil {
		fmt.Println(err)
	} else {
		logs := []model.JobLog{}
		err = dynamodbattribute.UnmarshalListOfMaps(resp1.Items, &logs)
		if err != nil {
			t.Fatal(err)
		}
		data, err := json.Marshal(logs)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(data))
	}
}

func TestTimeDuration(t *testing.T) {
	now := time.Now()

	t.Log(now.Add((-1) * time.Hour))
	d := DurationInput{
		Count:     1,
		Duration:  Weeks,
		Direction: Before,
	}

	t.Log(now.Add(d.GetDuration()))

	d = DurationInput{
		Count:     1,
		Duration:  Weeks,
		Direction: After,
	}

	t.Log(now.Add(d.GetDuration()))
}
