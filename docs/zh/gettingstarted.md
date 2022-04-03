# 快速开始

## 视频教程

[这里](https://www.bilibili.com/video/BV1k5411L79A/) 有一个视频教程，欢迎观看。

## 安装

你可以通过以下三种方式安装 `tape`：

1. **下载二进制文件**：从[这里](https://github.com/hyperledger-twgc/tape/releases)下载 tar 包，并解压。

2. **本地编译**：克隆本仓库并在根目录运行如下命令进行编译：

    ```
    make tape
    ```
 
    注意：

    1. 推荐 Go 1.18。Go 语言的安装请参考[这里](https://golang.google.cn/doc/install)。

    2. Tape 项目是一个 go module 工程，因此不用将项目保存到 `GOPATH` 下，任意目录都可执行编译操作。执行编译命令之后，它会自动下载相关依赖，下载依赖可能需要一定时间。编译完成后，会在当前目录生成一个名为 tape 的可执行文件。

    3. 如果下载依赖速度过慢，推荐配置 goproxy 国内代理，配置方式请参考[Goproxy 中国](https://goproxy.cn/)。

3. **拉取 Docker 镜像**: 

```shell
docker pull ghcr.io/hyperledger-twgc/tape
```

## 编译 Docker（可选）

Tape 支持本地编译 Docker 镜像，在项目根目录下执行以下命令即可：

```shell
make docker
```

执行成功之后本地会增加一个 ghcr.io/hyperledger-twgc/tape:latest 的 Docker 镜像。

## 修改配置文件

请根据[配置文件说明](configfile.md)修改配置文件。
注意：如果需要修改 hosts 文件，请注意相关映射的修改。

## 运行
### 默认模式
执行如下命令即可运行测试：
```shell
./tape --config=config.yaml --number=40000
```
如果需要使用其他模式，请参考：
### CommitOnly
```shell
docker run -v $PWD:/tmp ghcr.io/hyperledger-twgc/tape tape commitOnly -c $CONFIG_FILE -n 40000
```
### EndorsementOnly
```shell
docker run -v $PWD:/tmp ghcr.io/hyperledger-twgc/tape tape endorsementOnly -c $CONFIG_FILE -n 40000
```

该命令的含义是，使用 config.yaml 作为配置文件，向 Fabric 网络发送40000条交易进行性能测试。 
> 使用 `./tape --help` 可以查看 tape 帮助文档

注意：**请把发送交易数量设置为 batchsize （Fabric 中 Peer 节点的配置文件 core.yaml 中的参数，表示区块中包含的交易数量）的整倍数，这样最后一个区块就不会因为超时而出块了。** 例如，如果你的区块中包含交易数设为500，那么发送交易数量就应该设为1000、40000、100000这样的值。


## 日志说明

我们使用 [logrus](https://github.com/sirupsen/logrus) 来管理日志，请通过环境变量 `TAPE_LOGLEVEL` 来设置日志级别。例如：

```shell
export TAPE_LOGLEVEL=debug
```

日志级别共有如下八级，默认级别为 `warn`：
- panic
- fatal
- error
- warn
- warning
- info
- debug
- trace

## 注意事项

- 请把 Tape 和 Fabric 部署在接近的位置，或者直接部署在同一台机器上。这样就可以防止因网络带宽问题带来的瓶颈。你可以使用类似 `iftop` 这样的工具来监控网络流量。
- 可以使用类似 `top` 这样的指令查看 CPU 状态。在测试刚开始的时候，你可能会看到 CPU 使用率很高，这是因为 Peer 节点在处理提案。这种现象出现的时间会很短，然后你就会看到区块一个接一个的被提交。
- 修改 Fabric 的出块参数，可能会有不同的测试结果。如果你想测试最佳出块参数，请查看 [Probe](https://github.com/SamYuan1990/Probe) 项目，该项目的目的就是测试 Fabric 的最佳出块参数。
