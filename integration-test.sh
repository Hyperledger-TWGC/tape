#!/usr/bin/env bash
set -e

DIR=$PWD

curl -vsS https://raw.githubusercontent.com/hyperledger/fabric/master/scripts/bootstrap.sh | bash
cd ./fabric-samples/first-network
echo y | ./byfn.sh up
cp -r crypto-config "$DIR" && cd "$DIR"
go build
STUPID_LOGLEVEL=debug ./stupid config.json 100
