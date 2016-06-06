#!/bin/bash
. config-default.sh
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

docker build -t $imageName .

docker rm -f $containerName

docker run --link short-url-redis:redis -p 8001:8001 --name $containerName -d $imageName
