#!/usr/bin/env bash
git submodule update --init --recursive
cd fabric-samples
curl -vsS https://raw.githubusercontent.com/hyperledger/fabric/master/scripts/bootstrap.sh -o ./scripts/bootstrap.sh
chmod +x ./scripts/bootstrap.sh
./scripts/bootstrap.sh
cd ./first-network
sed -i "" 's/^askProceed/#askProceed/g' byfn.sh
./byfn.sh up
cd ../..
cp ./fabric-samples/first-network/crypto-config .
go build
./stupid config.json 10