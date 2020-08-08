#!/usr/bin/env bash
set -ex

DIR=$PWD
go build ./cmd/stupid
curl -vsS https://raw.githubusercontent.com/hyperledger/fabric/master/scripts/bootstrap.sh | bash

cd ./fabric-samples/
case $1 in
 14)
    git checkout release-1.4
    cd first-network
    echo y | ./byfn.sh up -i 1.4.8
    # comments here for 1.4.8 work around as docker image issue.
    # docker pull hyperledger/fabric-orderer:amd64-1.4  
    # Error response from daemon: manifest for hyperledger/fabric-orderer:amd64-1.4 not found: manifest unknown: manifest unknown
    cp -r crypto-config "$DIR"
    
    CONFIG_FILE=./test/config14org1andorg2.yaml
    ;;
 22)
    cd ./test-network
    echo y |  ./network.sh down -i 2.2
    echo y |  ./network.sh up createChannel -i 2.2
    cp -r organizations "$DIR"

    case $2 in
      ORLogic)
         CONFIG_FILE=./test/config20selectendorser.yaml
         ARGS=(-ccep "OR('Org1.member','Org2.member')")
         ;;
      ENDORSEMNTONLY)
         CONFIG_FILE=./test/config20EndorsementOnly.yaml
         ARGS=(-ccep "OR('Org1.member','Org2.member')")
         ;;
      COMMITONLY)
         CONFIG_FILE=./test/config20CommiterOnly.yaml
         ARGS=(-ccep "OR('Org1.member','Org2.member')")
         ;;
      *)
         CONFIG_FILE=./test/config20org1andorg2.yaml
         ;;
      esac
    echo y |  ./network.sh deployCC "${ARGS[@]}"
    ;;
 *)
    echo "Usage: $1 [14|22]"
    echo "When given version, start byfn or test network basing on specific version of docker image"
    echo "For any value without mock, 14, 22 will show this hint"
    exit 0
    ;;
esac

cd "$DIR"
STUPID_LOGLEVEL=debug ./stupid $CONFIG_FILE 100
