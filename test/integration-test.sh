#!/usr/bin/env bash
set -ex

DIR=$PWD
go build ./cmd/stupid
curl -vsS https://raw.githubusercontent.com/hyperledger/fabric/master/scripts/bootstrap.sh | bash

cd ./fabric-samples/test-network
echo y |  ./network.sh down
echo y |  ./network.sh up createChannel
cp -r organizations "$DIR"

CONFIG_FILE=./test/configorg1andorg2.yaml

if [ $1 == "ORLogic" ]; then
  CONFIG_FILE=./test/configselectendorser.yaml
  ARGS=(-ccep "OR('Org1.member','Org2.member')")
fi

echo y |  ./network.sh deployCC "${ARGS[@]}"
cd "$DIR"
STUPID_LOGLEVEL=debug ./stupid $CONFIG_FILE 100
