# 工作流程

Tape 有多个工作协程组成，所以该流程是高度并行化且可扩展的。这些协程通过缓存通道互相连接，所以它们可以互相传递数据。


Tape consists of several workers that run in goroutines, so that the pipeline is highly concurrent and scalable. Workers are connected via buffered channels, so they can pass products around.

整体工作流程如下图：

![tape workflow](images/tape.jpeg)

- **Signer**，签名交易提案协程，负责签名生成的交易提案，并将签名后的结果存入缓存通道中；
- **Proposer**，提案发送线程，负责从缓存通道中取出已签名的交易提案，然后通过 gRPC 将已签名提案发送到背书节点，并将背书节点返回的背书结果写入另一个缓存通道；
- **Integrator**，负责从缓存通道中取出背书后的结果，并封装成信封，信封是排序节点可接受的格式，然后将该信封再次存入一个单独的缓存通道；
- **Broadcaster**，负责将从缓存通道中取出信封，并然后通过 gRPC 将信封广播到排序节点；

以上四个协程可以启动不止一个，因此 Tape 实现高性能和可扩展性，Tape 自身不会成为性能瓶颈。

排序节点生成区块后，会将区块广播到 Peer 节点，Peer 节点接收到区块并经过验证保存到本地账本之后，会向其他节点广播已提交区块，

- **Observer**，接收到 Peer 节点广播的区块之后，会计算区块中交易数量，以及总耗时，当接收到区块的交易数量和运行 Tape 时输入的参数一致时，结束运行，并根据总耗时计算 TPS。

