# Tape
<div align="center">
<img src="logo.svg" width="100">
</div>
Tape 是一款轻量级 Hyperledger Fabric 性能测试工具

[![Go doc](https://img.shields.io/badge/go.dev-reference-brightgreen?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/hyperledger-twgc/tape)
[![Github workflow test](https://github.com/Hyperledger-TWGC/tape/actions/workflows/test.yml/badge.svg)](https://github.com/Hyperledger-TWGC/tape/actions/workflows/test.yml)

## 项目背景

Tape 项目原名 Stupid，最初由 超级账本中国技术工作组成员[郭剑南](https://github.com/guoger)开发，目的是提供一款轻量级、可以快速测试 Hyperledger Fabric TPS 值的工具。Stupid 取自[KISS](https://en.wikipedia.org/wiki/KISS_principle) 原则 Keep it Simple and Stupid，目前已正式更名为Tape，字面含义卷尺，寓意测量，测试。

目前 Tape 已贡献到超级账本中国技术社区，由[TWGC 性能优化小组](https://github.com/Hyperledger-TWGC/fabric-performance-wiki)负责维护。

## 项目特点

1. **轻量级**， Tape 实现过程中没有使用 SDK，直接使用 gRPC 向 Fabric 节点发送和接收请求；
2. **易操作**，通过简单的配置文件和命令即可快速启动测试；
3. **结果准确**，Tape 直接使用 gRPC 发送交易，并且对交易和区块处理的不同阶段单独拆分，使用协程及通道缓存的方式并行处理，大幅度提升了 Tape 自身的处理效率，从而可以准确的测试出 Fabric 的真实性能。
4. **参考标准** 其设计和功能参考[性能测试白皮书](https://github.com/Hyperledger-TWGC/fabric-performance-wiki/blob/master/performance-whitepaper.md)。

Tape由负载生成器客户端和观察者客户端组成。因此Tape仅可以用来对已经完成部署的Fabric网络进行测试。
- 负载生成器客户端
  - 直接使用了GRPC链接到被测网络而不使用任何SDK。因此避免了connection profile的配置， 减少了SDK的其他功能，如服务发现，可能带来的性能损耗。
- 观察者客户端会观察在多个peer节点上的提交，但不会进行资源的实时监控。

## 文档索引

如果你想快速使用 Tape 测试 TPS，请参考[快速开始](docs/zh/gettingstarted.md)；

如果你想了解配置文件中各项参数的具体含义，请参考[配置文件说明](docs/zh/configfile.md)；

如果你想详细了解 Tape 工作流程，请参考[工作流程](docs/zh/workflow.md)；

如果你在使用过程中遇到了问题请参考[FAQ](https://github.com/Hyperledger-TWGC/tape/wiki/FAQ)，如果 FAQ 还不能解决你的问题，请在 github 中提 issue，或者发邮件咨询项目维护者。


## [如何贡献](CONTRIBUTING.md)

## [维护者信息](MAINTAINERS.md)

## 使用许可

Tape 遵守 [Apache 2.0 开源许可](LICENSE)。

## Credits
Icons made by <a href="https://www.flaticon.com/authors/good-ware" title="Good Ware">Good Ware</a> from <a href="https://www.flaticon.com/" title="Flaticon">www.flaticon.com</a>