#!/bin/bash
. config-default.sh
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

docker build -t $imageName . && docker push $imageName
