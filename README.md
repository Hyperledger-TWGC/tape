# A Stupid traffic generator for Fabric

## Why Stupid

Sometimes we need to test performance of a deployed Fabric network with ease. There are many excellent projects out there, i.e. Hyperledger Caliper. However, we sometimes just need a tiny, handy tool, like `Stupid`.

About the name: [Keep It Simple, Stupid](https://en.wikipedia.org/wiki/KISS_principle)

## What is it

This is a very stupid traffic generator:
- it does not use any SDK
- it does not attempt to deploy Fabric
- it does not rely on connection profile
- it does not discover nodes, chaincodes, or policies
- it does not monitor resource utilization

It is used to perform super simple performance test:
- it directly establishes number of gRPC connections
- it sends signed proposals to peers via number of gRPC clients
- it assembles endorsed responses into envelopes
- it sends envelopes to orderer
- it observes transaction commitment

This tool is so stupid that *it will not be the bottleneck of performance test*

## How to use it

### Make sure you have Go 1.12 installed

### clone this repo and run `go build` at root dir
This is a go module project so you don't need to clone it into `GOPATH`. It will download required dependencies automatically, which may take a while depending on network connection. Once it finishes building, you should have a executable named `stupid`.

### modify `config.json` according to your network
This is a sample:
```json
{
  "peer_addr": "peer0.org1.example.com:7051",
  "orderer_addr": "orderer.example.com:7050",
  "channel": "mychannel",
  "chaincode": "mycc",
  "args": ["put", "key", "value"],
  "mspid": "Org1MSP",
  "private_key": "wallet/priv.key",
  "sign_cert": "wallet/sign.crt",
  "tls_ca_certs": ["wallet/pca.crt","wallet/oca.crt"],
  "num_of_conn": 20,
  "client_per_conn": 40
}
```

`peer_addr`: peer address in IP:Port format. It does not support sending traffic to multiple peers, yet. You may need to add peer name, i.e. `peer0.org1.example.com` to your `/etc/hosts`

`orderer_addr`: orderer address in IP:Port format. It does not support sending traffic to multiple orderers, yet. You may need to add orderer name, i.e. `orderer.example.com` to your `/etc/hosts`

This tool sends traffic as a Fabric user, and requires following configs

`mspid`: MSP ID that the user is associated to

`private_key`: path to the private key. If you are using BYFN as your base, this can be `crypto-config/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/keystore/priv_sk`

`sign_cert`: path to the user certificate. If you are using BYFN as your base, this can be `crypto-config/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/signcerts/User1@org1.example.com-cert.pem`

`tls_ca_certs`: this contains TLS CA certificates of peer and orderer. If tls is disabled, leave this blank. Otherwise, it can be `crypto-config/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem` from peer and `crypto-config/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem` from orderer

`channel`: channel name

`chaincode`: chaincode to invoke

`args`: arguments to send with invocation, depending on your chaincode implementation. The chaincode used by this sample can be found in `chaincodes/sample.go`

`num_of_conn`: number of gRPC connection established between client/peer, client/orderer. If you think client has not put enough pressure on Fabric, increase this.

`client_per_conn`: number of clients per connection used to send proposals to peer. If you think client has not put enough pressure on Fabric, increase this.

### Run `stupid config.json 40000` to generate 40000 transactions to Fabric.

*Set this to integer times of batchsize, so that last block is not cut due to timeout*. For example, if you have batch size of 500, set this to 500, 1000, 40000, 100000, etc.

## Tips

- Put this generator closer to Fabric, on even on the same machine. This is to prevent network bandwidth from being the bottleneck. You can use tools like `iftop` to monitor network traffic.

- Observe cpu status of peer with tools like `top`. You should see CPUs being exhausted at the beginning of test, that is peer processing proposals. It should be fairly quick, then you'll see blocks are being committed one after another.

- Increase number of messages per block in your channel configuration may help 
