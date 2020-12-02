#!/bin/bash

if [ -z "$QUEUE_NAME" ]; then
    echo "Miss ENV QUEUE_NAME"
    exit 1
fi


/app/controlplane  -addr :80 -q ${QUEUE_NAME} -verbose