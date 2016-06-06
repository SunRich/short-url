#!/bin/sh
. config-default.sh

ssh root@$KYYTESTSERVER << EOF

docker pull $imageName
docker stop $containerName
docker rm $containerName
docker run -p 8002:8001 \
           -e REDIS_PORT_6379_TCP_ADDR=192.168.42.1 \
           --name $containerName -d $imageName
EOF
