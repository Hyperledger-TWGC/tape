#!/usr/bin/env bash
set -ex

DIR=$PWD

curl -vsS https://raw.githubusercontent.com/hyperledger/fabric/master/scripts/bootstrap.sh | bash
cd ./fabric-samples/test-network
echo y |  ./network.sh up createChannel
echo y |  ./network.sh deployCC
cp -r organizations "$DIR" && cd "$DIR"
go build ./cmd/stupid
STUPID_LOGLEVEL=debug ./stupid config.yaml 100


