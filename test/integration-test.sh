#!/usr/bin/env bash
set -ex

DIR=$PWD
go build ./cmd/stupid
curl -vsS https://raw.githubusercontent.com/hyperledger/fabric/master/scripts/bootstrap.sh | bash

cd ./fabric-samples/
if [ $1 == "1.4" ]; then
  git checkout release-1.4
  cd first-network
  echo y | ./byfn.sh up -i 1.4.8
  # comments here for 1.4.8 work around as docker image issue.
  # docker pull hyperledger/fabric-orderer:amd64-1.4  
  # Error response from daemon: manifest for hyperledger/fabric-orderer:amd64-1.4 not found: manifest unknown: manifest unknown
  cp -r crypto-config "$DIR"
  
  CONFIG_FILE=./test/config14org1andorg2.yaml
fi

if [ $1 == "2.2" ]; then
  cd ./test-network
  echo y |  ./network.sh down -i $1
  echo y |  ./network.sh up createChannel -i $1
  cp -r organizations "$DIR"

  CONFIG_FILE=./test/config20org1andorg2.yaml

  if [ $2 == "ORLogic" ]; then
    CONFIG_FILE=./test/config20selectendorser.yaml
    ARGS=(-ccep "OR('Org1.member','Org2.member')")
  fi

  echo y |  ./network.sh deployCC "${ARGS[@]}"
fi

cd "$DIR"
STUPID_LOGLEVEL=debug ./stupid $CONFIG_FILE 100
