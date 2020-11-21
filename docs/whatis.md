# 介绍

Tape 一款轻量级 Hyperledger Fabric 性能测试工具。

## 项目背景

Tape 项目原名 Stupid，最初由 [TWGC（Technical Working Group China，超级账本中国技术工作组）](https://wiki.hyperledger.org/display/TWGC)成员郭剑南开发，目的是提供一款轻量级、可以快速测试 Hyperledger Fabric TPS 值的工具。Stupid 取自 [KISS](https://en.wikipedia.org/wiki/KISS_principle) 原则 Keep it Simple and Stupid，目前已正式更名为 Tape，字面含义卷尺，寓意测量，测试。目前 Tape 已贡献到超级账本中国技术社区，由 [TWGC 性能优化小组](https://github.com/Hyperledger-TWGC/fabric-performance-wiki)负责维护。

## 它不做什么

- 它不使用任何 SDK
- 它不会尝试部署 Fabric 网络
- 它不会发现节点、链码或者策略
- 它不会监控资源使用

## 项目特点

1. **轻量级**， Tape 实现过程中没有使用 SDK，直接使用 gRPC 向 Fabric 节点发送和接收请求；
2. **易操作**，通过简单的配置文件和命令即可快速启动测试；
3. **结果准确**，Tape 直接使用 gRPC 发送交易，并且对交易和区块处理的不同阶段单独拆分，使用协程及通道缓存的方式并行处理，大幅度提升了 Tape 自身的处理效率，从而可以准确的测试出 Fabric 的真实性能。

## 文档阅读指南

如果你想快速使用 Tape 测试 TPS，请参考[快速开始](gettingstarted.md)；
如果你想了解配置文件中各项参数的具体含义，请参考[配置文件说明](configfile.md)；
如果你想详细了解 Tape 工作流程，请参考[工作流程](workflow.md)；
如果你在使用过程中遇到了问题请参考[FAQ](FAQ.md)，如果 FAQ 还不能解决你的问题，请在 github 中提 issue，或者发邮件咨询项目维护者。

## 欢迎贡献

如果你希望提交新的特性或者遇到了任何 Bug，欢迎在 github 仓库中开启新的 [issue](https://github.com/guoger/tape/issues)，同时也欢迎提交 [pull request](https://github.com/guoger/tape/pulls)。

## 贡献者信息

| 姓名   | 邮箱                     | github-ID   | 所属组织                                          | 角色   |
| ------ | ------------------------ | ----------- | ------------------------------------------------- | ------ |
| 郭剑南 | guojiannan1101@gmail.com | guoger      | [TWGC](https://wiki.hyperledger.org/display/TWGC) | 维护者 |
| 袁怿   | yy19902439@126.com       | SamYuan1990 | [TWGC](https://wiki.hyperledger.org/display/TWGC) | 维护者 |
| 程阳   | chengyang418@163.com     | stone-ch    | [TWGC](https://wiki.hyperledger.org/display/TWGC) | 维护者 |

## 使用许可

Tape 遵守 Apache 2.0 开源许可。
