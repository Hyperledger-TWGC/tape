# Configuraion in Details

Tape need a configuration file as `config.yaml`. You can find it in project root. Before start Tape to test your own network, please modify it accordingly.
This is a sample:
```yaml
# Definition of nodes
peer1: &peer1
  addr: localhost:7051
  ssl_target_name_override: peer0.org1.example.com
  org: org1
  tls_ca_cert: /config/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/msp/tlscacerts/tlsca.org1.example.com-cert.pem

peer2: &peer2
  addr: localhost:9051
  ssl_target_name_override: peer0.org2.example.com
  org: org2
  tls_ca_cert: /config/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/msp/tlscacerts/tlsca.org2.example.com-cert.pem

orderer1: &orderer1
  addr: localhost:7050
  ssl_target_name_override: orderer.example.com
  org: org1
  tls_ca_cert: /config/crypto-config/ordererOrganizations/example.com/msp/tlscacerts/tlsca.example.com-cert.pem

policyFile: /config/test/andLogic.rego

# Nodes to interact with
endorsers:
  - *peer1
# we might support multi-committer in the future for more complex test scenario,
# i.e. consider tx committed only if it's done on >50% of nodes. But for now,
# it seems sufficient to support single committer.
committers: 
  - *peer1
  - *peer2

commitThreshold: 1

orderer: *orderer1

# Invocation configs
channel: mychannel
chaincode: basic
args:
  - CreateAsset
  - uuid
  - randomString8
  - randomNumber0_50
  - randomString8
  - randomNumber0_50
mspid: Org1MSP
private_key: /config/organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/keystore/priv_sk
sign_cert: /config/organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/signcerts/User1@org1.example.com-cert.pem
num_of_conn: 10
client_per_conn: 10
```
Let's deep dive the config.
1st node related setting:
```yaml
# Definition of nodes
peer1: &peer1
  ...

peer2: &peer2
  ...

orderer1: &orderer1
  ...
```

Here defines for nodes, including peer and orderer. we need address, org name (for endorsement policy usage), and (m)TLS certs if any.
```yaml
peer1: &peer1
  addr: localhost:7051
  org: org1
  tls_ca_cert: ./organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/msp/tlscacerts/tlsca.org1.example.com-cert.pem
  tls_ca_key: ./organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.key
  tls_ca_root: ./organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.crt

peer2: &peer2
  addr: localhost:9051
  org: org2
  tls_ca_cert: ./organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/msp/tlscacerts/tlsca.org2.example.com-cert.pem
  tls_ca_key: ./organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/server.key
  tls_ca_root: ./organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/server.crt

orderer1: &orderer1
  addr: localhost:7050
  org: org1
  tls_ca_cert: ./organizations/ordererOrganizations/example.com/orderers/orderer0.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
  tls_ca_key: ./organizations/ordererOrganizations/example.com/orderers/orderer0.example.com/tls/server.key
  tls_ca_root: ./organizations/ordererOrganizations/example.com/orderers/orderer0.example.com/tls/server.crt
```

- `tls_ca_cert` TLS cert。
- `tls_ca_key`：TLS private key。
- `tls_ca_root`：CA root cert。

Then move to endorsement and commit parts:

```yaml
policyFile: /config/test/andLogic.rego
# Nodes to interact with
endorsers:
  - *peer1
  - *peer2
# we support multi-committer in the future for more complex test scenario,
# i.e. consider tx committed only if it's done on >50% of nodes. But for now,
# it seems sufficient to support single committer.
committers: 
  - *peer1
  - *peer2

commitThreshold: 1
orderer: *orderer1
```

We defined endorsement peer, commit peer and orderer node in each sections. With `policyFile` for given endorsement policy. So far we use OPA and rego for endorsement policy. You can file as sample for a policy as `org1` and `org2`  can be described as below:
```
package tape

default allow = false
		
allow {
  input[_] == "org1"
  input[_] == "org2"
}
```

`endorsers`: peer node for endorsement the traffic
  - include the addr and tls ca cert of peers. Peer address is in IP:Port format. 
  - You may need to add peer name, i.e. `peer0.org1.example.com,peer0.org2.example.com` to your `/etc/hosts`
`committers`: peer node for block commit
  - observe tx commitment from these peers. If you want to observe over 50% of peers on your network, you should selected and put them here.
`commitThreshold`: how many peers committed the block as the block logically committed among fabric network.
  - mark the block as successe, after how many committer receive the mesage
`orderer`: orderer node
  - include the addr and tls ca cert of orderer. Orderer address is in IP:Port format. It does not support sending traffic to multiple orderers, yet. 
  - You may need to add orderer name, i.e. `orderer.example.com` to your `/etc/hosts`

As Tape sends traffic as a Fabric user, and requires following configs

```yaml
# Invocation configs
channel: mychannel
chaincode: basic
args:
  - CreateAsset
  - uuid
  - randomString8
  - randomNumber0_50
  - randomString8
  - randomNumber0_50
mspid: Org1MSP
private_key: ./organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/keystore/priv_sk
sign_cert: ./organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/signcerts/User1@org1.example.com-cert.pem
num_of_conn: 10
client_per_conn: 10
```

`channel`：channel name

`chaincode`: chaincode name

`version`: the version of chaincode. This is left to empty by default.

`args`：args for chaincode action, take sample from [abac](https://github.com/hyperledger/fabric-samples/blob/master/chaincode/abac/go/abac.go) ，if from alice trans 10 to bob.

```
args:
  - invoke
  - a
  - b
  - 10
```

for random arg support, we support `uuid`,`randomString$length`，`randomNumberA_B`
```
args:
  - CreateAsset
  - uuid
  - randomString8
  - randomNumber0_50
  - randomString8
  - randomNumber0_50
```

`mspid`：MSP ID.

`private_key`：private key for user, sample as `crypto-config/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/keystore/priv_sk` 。

`sign_cert`： cert for user, sample as `crypto-config/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/signcerts/User1@org1.example.com-cert.pem` 。

`num_of_conn`：over all connection setting between tape client and peer, tape client and orderer

`client_per_conn`：connection number between tape client and peer, `Total connections number = num_of_conn * client_per_conn`