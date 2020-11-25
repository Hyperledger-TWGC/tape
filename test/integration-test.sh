#!/usr/bin/env bash
set -ex

DIR=$PWD
docker build -t tape:latest .

case $1 in
 1_4)
    # sadly, bootstrap.sh from release-1.4 still pulls binaries from Nexus, which is not available anymore
    curl -vsS https://raw.githubusercontent.com/hyperledger/fabric/release-2.2/scripts/bootstrap.sh | bash
    cd ./fabric-samples/
    git checkout release-1.4
    cd ./first-network
    echo y | ./byfn.sh up -i 1.4.8
    # comments here for 1.4.8 work around as docker image issue.
    # docker pull hyperledger/fabric-orderer:amd64-1.4  
    # Error response from daemon: manifest for hyperledger/fabric-orderer:amd64-1.4 not found: manifest unknown: manifest unknown
    cp -r crypto-config "$DIR"
    
    CONFIG_FILE=/config/test/config14org1andorg2.yaml
    ;;
 2_2)
    curl -vsS https://raw.githubusercontent.com/hyperledger/fabric/release-2.2/scripts/bootstrap.sh | bash
    cd ./fabric-samples/test-network
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
 latest)
    curl -vsS https://raw.githubusercontent.com/hyperledger/fabric/master/scripts/bootstrap.sh | bash
    cd ./fabric-samples/test-network
    echo y |  ./network.sh down
    echo y |  ./network.sh up createChannel
    cp -r organizations "$DIR"

    CONFIG_FILE=/config/test/config20org1andorg2.yaml

    if [ $2 == "ORLogic" ]; then
      CONFIG_FILE=/config/test/config20selectendorser.yaml
      ARGS=(-ccep "OR('Org1.member','Org2.member')")
    fi

    echo y |  ./network.sh deployCC "${ARGS[@]}"
    ;;
 *)
    echo "Usage: $1 [1_4|2_2|latest]"
    echo "When given version, start byfn or test network basing on specific version of docker image"
    echo "For any value without mock, 1_4, 2_2, latest will show this hint"
    exit 0
    ;;
esac

cd "$DIR"
docker run  -e TAPE_LOGLEVEL=debug --network host -v $PWD:/config tape tape $CONFIG_FILE 500
