#!/usr/bin/env bash
set -e
#git submodule update --init --recursive
#git clone https://github.com/hyperledger/fabric-samples.git
#cd fabric-samples
DIR=$PWD
curl -vsS https://raw.githubusercontent.com/hyperledger/fabric/master/scripts/bootstrap.sh | bash
cd ./fabric-samples/first-network
echo y | ./byfn.sh up
cp -r crypto-config "$DIR" && cd "$DIR"
go build
./stupid config.json 10