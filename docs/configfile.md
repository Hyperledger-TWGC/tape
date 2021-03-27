# 配置文件说明

我们为 Tape 提供了一个示例配置文件 `config.yaml`，你可以在项目根目录下找到它。使用 Tape 进行测试之前，请根据您的区块链网络情况修改该配置文件。
Modify `config.yaml` according to your network

`config.yaml` 示例配置文件如下所示：

This is a sample:

```yaml
# Definition of nodes
peer1: &peer1
  addr: localhost:7051
  tls_ca_cert: ./organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/msp/tlscacerts/tlsca.org1.example.com-cert.pem

peer2: &peer2
  addr: localhost:9051
  tls_ca_cert: ./organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/msp/tlscacerts/tlsca.org2.example.com-cert.pem

orderer1: &orderer1
  addr: localhost:7050
  tls_ca_cert: ./organizations/ordererOrganizations/example.com/msp/tlscacerts/tlsca.example.com-cert.pem

# Nodes to interact with
endorsers:
  - *peer1
  - *peer2
# we might support multi-committer in the future for more complex test scenario,
# i.e. consider tx committed only if it's done on >50% of nodes. But for now,
# it seems sufficient to support single committer.
committer: *peer2
orderer: *orderer1

# Invocation configs
channel: mychannel
chaincode: basic
args:
  - GetAllAssets
mspid: Org1MSP
private_key: ./organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/keystore/priv_sk
sign_cert: ./organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/signcerts/User1@org1.example.com-cert.pem
num_of_conn: 10
client_per_conn: 10

```
接下来我们将逐一解析该配置文件的含义。

首先，前三个部分：

```yaml
# Definition of nodes
peer1: &peer1
  addr: localhost:7051
  tls_ca_cert: ./organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/msp/tlscacerts/tlsca.org1.example.com-cert.pem

peer2: &peer2
  addr: localhost:9051
  tls_ca_cert: ./organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/msp/tlscacerts/tlsca.org2.example.com-cert.pem

orderer1: &orderer1
  addr: localhost:7050
  tls_ca_cert: ./organizations/ordererOrganizations/example.com/msp/tlscacerts/tlsca.example.com-cert.pem
```

定义了不同的节点，包括 Peer 节点和排序节点，配置中需要确认节点地址以及 TLS CA 证书（如果启用 TLS，则必须配置 TLS CA 证书）。其中节点地址格式为`地址:端口`。此处`地址`推荐使用域名，因此您可能还需要在 hosts 文件中增加节点域名和 IP 的映射关系。

如果启用了双向 TLS，即你的 Fabric 网络中的 Peer 节点在 core.yaml 配置了 "peer->tls->clientAuthRequired" 为 "true"，则表明，不但服务端（Peer 节点）向客户端（Tape）发送的信息是经过加密的，客户端（Tape）向服务端（Peer 节点）发送的信息也应该是加密的，因此我们就需要在配置文件中增加 TLS 通信中需要使用的密钥，双向 TLS 配置示例如下：

```yaml
peer1: &peer1
  addr: localhost:7051
  tls_ca_cert: ./organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/msp/tlscacerts/tlsca.org1.example.com-cert.pem
  tls_ca_key: ./organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.key
  tls_ca_root: ./organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.crt

peer2: &peer2
  addr: localhost:9051
  tls_ca_cert: ./organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/msp/tlscacerts/tlsca.org2.example.com-cert.pem
  tls_ca_key: ./organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/server.key
  tls_ca_root: ./organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/server.crt

orderer1: &orderer1
  addr: localhost:7050
  tls_ca_cert: ./organizations/ordererOrganizations/example.com/orderers/orderer0.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
  tls_ca_key: ./organizations/ordererOrganizations/example.com/orderers/orderer0.example.com/tls/server.key
  tls_ca_root: ./organizations/ordererOrganizations/example.com/orderers/orderer0.example.com/tls/server.crt
```

其中三个 TLS 相关的证书/密钥说明如下：
- `tls_ca_cert`：客户端 TLS 通信时使用的证书文件。
- `tls_ca_key`：客户端 TLS 通信时使用的私钥文件。
- `tls_ca_root`：CA 根证书文件。

接下来的三个部分：

```yaml
# Nodes to interact with
endorsers:
  - *peer1
  - *peer2
# we might support multi-committer in the future for more complex test scenario,
# i.e. consider tx committed only if it's done on >50% of nodes. But for now,
# it seems sufficient to support single committer.
committer: *peer2
orderer: *orderer1
```

分别定义了角色为背书节点（endorsers）、提交节点（committer）和排序节点（orderer）的节点。

`endorsers`: 负责为交易提案背书的节点，Tape 会把构造好的已签名的交易提案发送到背书节点进行背书。
  - include the addr and tls ca cert of peers. Peer address is in IP:Port format. 
  - You may need to add peer name, i.e. `peer0.org1.example.com,peer0.org2.example.com` to your `/etc/hosts`
`committer`: 负责接收其他节点广播的区块提交成功的信息。
  - observe tx commitment from these peers. If you want to observe over 50% of peers on your network, you should selected and put them here.
`orderer`: 排序节点，目前 Tape 仅支持向一个排序节点发送交易排序请求。
  - include the addr and tls ca cert of orderer. Orderer address is in IP:Port format. It does not support sending traffic to multiple orderers, yet. 
  - You may need to add orderer name, i.e. `orderer.example.com` to your `/etc/hosts`
Tape 以 Fabric 用户的身份向区块链网络发送交易，所以还需要下边的配置：

This tool sends traffic as a Fabric user, and requires following configs

```yaml
# Invocation configs
channel: mychannel
chaincode: basic
args:
  - GetAllAssets
mspid: Org1MSP
private_key: ./organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/keystore/priv_sk
sign_cert: ./organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/signcerts/User1@org1.example.com-cert.pem
num_of_conn: 10
client_per_conn: 10
```

`channel`：通道名。

`chaincode`：要调用的链码名。

`commitThreshold`: how many committers received the block see as successed.

`version`: the version of chaincode. This is left to empty by default.

`args`：要调用的链码的参数。参数取决于链码实现，例如，fabric-samples 项目中提供的示例链码 [abac](https://github.com/hyperledger/fabric-samples/blob/master/chaincode/abac/go/abac.go) ，其功能为账户A和账户B之间的转账。如果想要以此链码作为性能测试的链码，执行操作为账户A向账户B转账10，则参数设置如下：

```
args:
  - invoke
  - a
  - b
  - 10
```

`mspid`：MSP ID 是用户属性的一部分，表明该用户所属的组织。

`private_key`：用户私钥的路径。如果你使用 BYFN 作为你的测试网络，私钥路径为 `crypto-config/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/keystore/priv_sk` 。

`sign_cert`：用户证书的路径。如果你使用 BYFN 作为你的测试网络，私钥路径为 `crypto-config/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/signcerts/User1@org1.example.com-cert.pem` 。

`num_of_conn`：客户端和 Peer 节点，客户端和排序节点之间创建的 gRPC 连接数量。如果你觉得向 Fabric 施加的压力还不够，可以将这个值设置的更大一些。

`client_per_conn`：每个连接用于向每个 Peer 节点发送 提案的客户端数量。如果你觉得向 Fabric 施加的压力还不够，可以将这个值设置的更大一些。所以 Tape 向 Fabric 发送交易的并发量为 `num_of_conn` * `client_per_conn`。