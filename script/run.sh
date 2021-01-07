#!/bin/bash

cd $(cd $(dirname $0); pwd)

while [ "$(ps -e > /tmp/psi && grep "proto_server\|proto_bridge" /tmp/psi)" != "" ]
do
    sleep 1
done

cd ../bin
env GOTRACEBACK=crash nohup ./proto_server > server.log 2>&1 &
env GOTRACEBACK=crash nohup ./proto_bridge > bridge.log 2>&1 &