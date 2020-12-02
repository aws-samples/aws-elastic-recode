#!/bin/bash

echo "PreStop processing"
queryURL=$(aws sqs get-queue-url --queue-name "$CONTROL_PLANE" --output text)
echo $queryURL
stopEvent=`echo  "{\"workerType\":\"${WORKER_TYPE}\",\"nodeName\":\"${WORKER_NODE_NAME}\",\"podName\": \"${WORKER_POD_NAME}\",\"action\": \"WorkerStop\",\"job\": \"{}\",\"msg\":\"pod termination\", \"ts\":$(date +%s)}"`
echo "$stopEvent"
aws sqs send-message --queue-url  $queryURL --message-body "$stopEvent"



