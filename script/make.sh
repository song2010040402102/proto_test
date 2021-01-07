#!/bin/bash

cd $(cd $(dirname $0); pwd)

if [ ! -d "../bin" ]; then
  mkdir ../bin
  cp ../config.yaml ../bin
fi

cd ../protobuf/protocol
./build.sh
cd ../../

if [[ $1 == "bridge" || $1 == "" ]]; then
    cd bridge
    go build -o ../bin/proto_bridge
fi
cd ..
if [[ $1 == "server" || $1 == "" ]]; then
    cd server
    go build -o ../bin/proto_server
fi