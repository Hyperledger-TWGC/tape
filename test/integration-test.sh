#!/usr/bin/env bash
set -ex

DIR=$PWD
docker build -t tape:latest .

curl -vsS https://raw.githubusercontent.com/hyperledger/fabric/master/scripts/bootstrap.sh | bash -s -- 2.2.0 1.4.7

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
    
    CONFIG_FILE=/config/test/config14org1andorg2.yaml
    ;;
 22)
    cd ./test-network
    echo y |  ./network.sh down -i 2.2
    echo y |  ./network.sh up createChannel -i 2.2
    cp -r organizations "$DIR"

    CONFIG_FILE=/config/test/config20org1andorg2.yaml

    if [ $2 == "ORLogic" ]; then
      CONFIG_FILE=/config/test/config20selectendorser.yaml
      ARGS=(-ccep "OR('Org1.member','Org2.member')")
    fi

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
docker run  -e TAPE_LOGLEVEL=debug --network host -v $PWD:/config tape tape $CONFIG_FILE 500
