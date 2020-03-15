#!/usr/bin/env bash
#git submodule update --init --recursive
git clone https://github.com/hyperledger/fabric-samples.git
cd fabric-samples
curl -vsS https://raw.githubusercontent.com/hyperledger/fabric/master/scripts/bootstrap.sh -o ./scripts/bootstrap.sh
chmod +x ./scripts/bootstrap.sh
./scripts/bootstrap.sh
cd ./first-network
sed -i 's/^askProceed/#askProceed/g' byfn.sh
./byfn.sh up
cd ../..
cp -r ./fabric-samples/first-network/crypto-config .
go build
./stupid config.json 10