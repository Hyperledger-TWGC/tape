#!/usr/bin/env bash
set -ex

DIR=$PWD
docker build -t tape:latest .
network=fabric_test
export COMPOSE_PROJECT_NAME=fabric
CMD=run

case $1 in
 1_4)
    unset COMPOSE_PROJECT_NAME
    # sadly, bootstrap.sh from release-1.4 still pulls binaries from Nexus, which is not available anymore
    # Why comment following code? Please check this issue: https://github.com/Hyperledger-TWGC/tape/issues/159
    # curl -vsS https://raw.githubusercontent.com/hyperledger/fabric/release-2.2/scripts/bootstrap.sh | bash
    ./test/bootstrap-v2.2.sh 1.4.12 1.5.2
    cd ./fabric-samples/
    git checkout release-1.4
    cd ./first-network

    echo y | ./byfn.sh up -i 1.4.12

    ## 1.4 cryptogen compensate
    priv_sk=crypto-config/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/keystore/*
    cp -f $priv_sk crypto-config/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/keystore/priv_sk
    ##
    cp -r crypto-config "$DIR"

    CONFIG_FILE=/config/test/config14org1andorg2.yaml
    network=host
    ;;
  2_2)
     curl -sSL https://bit.ly/2ysbOFE | bash -s -- 2.2.2 1.4.9
     cd ./fabric-samples/test-network
     echo y |  ./network.sh down -i 2.2
     echo y |  ./network.sh up createChannel -i 2.2
     cp -r organizations "$DIR"

     CONFIG_FILE=/config/test/config20org1andorg2.yaml

     case $2 in
      ORLogic)
         CONFIG_FILE=/config/test/configlatest.yaml
         ARGS=(-ccep "OR('Org1.member','Org2.member')")
         ;;
      ENDORSEMNTONLY)
         CONFIG_FILE=/config/test/configlatest.yaml
         ARGS=(-ccep "OR('Org1.member','Org2.member')")
         ;;
      COMMITONLY)
         CONFIG_FILE=/config/test/config20selectendorser.yaml
         ARGS=(-cci initLedger)
         ;;
      *)
         CONFIG_FILE=/config/test/configlatest.yaml
         ARGS=(-cci initLedger)
         ;;
      esac

     echo y |  ./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-go/ -ccl go "${ARGS[@]}"
     ;;
 2_5)
    curl -sSLO https://raw.githubusercontent.com/hyperledger/fabric/main/scripts/install-fabric.sh && chmod +x install-fabric.sh
    ./install-fabric.sh docker samples binary
    cd ./fabric-samples/test-network
    echo y |  ./network.sh down
    echo y |  ./network.sh up createChannel
    cp -r organizations "$DIR"

    CONFIG_FILE=/config/test/config20org1andorg2.yaml

    case $2 in
      ORLogic)
         CONFIG_FILE=/config/test/configlatest.yaml
         ARGS=(-ccep "OR('Org1.member','Org2.member')")
         ;;
      ENDORSEMNTONLY)
         CONFIG_FILE=/config/test/configlatest.yaml
         ARGS=(-ccep "OR('Org1.member','Org2.member')")
         ;;
      COMMITONLY)
         CONFIG_FILE=/config/test/config20selectendorser.yaml
         ARGS=(-cci initLedger)
         ;;
      *)
         CONFIG_FILE=/config/test/configlatest.yaml
         ARGS=(-cci initLedger)
         ;;
    esac

    echo y |  ./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-go/ -ccl go "${ARGS[@]}"
    ;;
 #latest)
 #   curl -vsS https://raw.githubusercontent.com/hyperledger/fabric/master/scripts/bootstrap.sh | bash
 #   cd ./fabric-samples/test-network
 #   echo y |  ./network.sh down
 #   echo y |  ./network.sh up createChannel
 #   cp -r organizations "$DIR"

    #CONFIG_FILE=/config/test/configlatest.yaml
    #ARGS=(-cci initLedger)

  #  case $2 in
  #    ORLogic)
  #       CONFIG_FILE=/config/test/configlatest.yaml
  #       ARGS=(-ccep "OR('Org1.member','Org2.member')")
  #       ;;
  #    ENDORSEMNTONLY)
  #       CONFIG_FILE=/config/test/configlatest.yaml
  #       ARGS=(-ccep "OR('Org1.member','Org2.member')")
  #       ;;
  #    COMMITONLY)
  #       CONFIG_FILE=/config/test/config20selectendorser.yaml
  #       ARGS=(-cci initLedger)
  #       ;;
  #    *)
  #       CONFIG_FILE=/config/test/configlatest.yaml
  #       ARGS=(-cci initLedger)
  #       ;;
  #  esac

  #  echo y |  ./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-go/ -ccl go "${ARGS[@]}"
   # ;;
 *)
    echo "Usage: $1 [1_4|2_2|2_5]"
    echo "When given version, start byfn or test network basing on specific version of docker image"
    echo "For any value without mock, 1_4,2_2,2_5 will show this hint"
    exit 0
    ;;
esac

cd "$DIR"
## warm up for the init chaincode block
sleep 10
case $2 in
      ORLogic)
         docker run -d --name tape3 -e TAPE_LOGLEVEL=debug --network $network -v $PWD:/config tape tape observer -c $CONFIG_FILE
         docker run -d --name tape1 -e TAPE_LOGLEVEL=debug --network $network -v $PWD:/config tape tape traffic -c $CONFIG_FILE --rate=10 -n 500
         docker run -d --name tape2 -e TAPE_LOGLEVEL=debug --network $network -v $PWD:/config tape tape traffic -c $CONFIG_FILE --rate=10 -n 500
         sleep 10
         timeout 10 docker logs tape3
         timeout 10 docker logs tape2
         ;;
      ENDORSEMNTONLY)
         CMD=endorsementOnly
         timeout 60 docker run --name tape -e TAPE_LOGLEVEL=debug --network $network -v $PWD:/config tape tape $CMD -c $CONFIG_FILE -n 500 --signers=10 --parallel=2
         ;;
      COMMITONLY)
         CMD=commitOnly
         timeout 60 docker run --name tape -e TAPE_LOGLEVEL=debug --network $network -v $PWD:/config tape tape $CMD -c $CONFIG_FILE -n 500 --signers=10 --parallel=2
         ;;
      *)
         docker run --name tape -e TAPE_LOGLEVEL=debug --network $network -v $PWD:/config tape tape $CMD -c $CONFIG_FILE -n 500
         ;;
esac

docker logs peer0.org1.example.com