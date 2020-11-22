# 快速开始

## 准备工作

Go 1.11 或者更高版本。推荐 Go 1.13。Go 语言的安装请参考[这里](https://golang.google.cn/doc/install)。

## 安装

目前，你需要从源码安装 Tape，且支持本地编译 Docker 镜像。 

使用如下命令克隆 Tape 仓库到本地：

```
git clone https://github.com/guoger/tape.git
```

如果已经配置 SSH 密钥，也可以使用如下命令克隆：

```
git clone git@github.com:guoger/tape.git
```

然后，执行如下命令，进入项目目录并编译：

```
cd tape
go build ./cmd/tape
```

Tape 项目是一个 go module 工程，因此不用将项目保存到 `GOPATH` 下，任意目录都可执行编译操作。执行编译命令之后，它会自动下载相关依赖，下载依赖可能需要一定时间。编译完成后，会在当前目录生成一个名为 tape 的可执行文件。

注意：如果下载依赖速度过慢，推荐配置 goproxy 国内代理，配置方式请参考[Goproxy 中国](https://goproxy.cn/)。

## 编译 Docker（可选）

Tape docker镜像下载
```shell
docker pull guoger/tape 
```
Tape 支持本地编译 Docker 镜像，在项目根目录下执行以下命令即可：

```shell
docker build -t guoger/tape:latest .
```

执行成功之后本地会增加一个 guoger/tape:latest 的 Docker 镜像。

## 修改配置文件

请根据[配置文件说明](configfile.md)修改配置文件。

注意：如果需要修改 hosts 文件，请注意相关映射的修改。

## 运行

执行如下命令即可运行测试：

```
./tape config.yaml 40000
```

该命令的含义是，使用 config.yaml 作为配置文件，向 Fabric 网络发送40000条交易进行性能测试。

注意：**请把发送交易数量设置为 batchsize （Fabric 中 Peer 节点的配置文件 core.yaml 中的参数，表示区块中包含的交易数量）的整倍数，这样最后一个区块就不会因为超时而出块了。** 例如，如果你的区块中包含交易数设为500，那么发送交易数量就应该设为1000、40000、100000这样的值。


## 日志说明

我们使用 [logrus](https://github.com/sirupsen/logrus) 来管理日志，请通过环境变量 `TAPE_LOGLEVEL` 来设置日志级别。例如：

```
export STUPID_LOGLEVEL=debug
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
- 修改 Peer 的出块参数，可能会有不同的测试结果。如果你想测试最佳出块参数，请查看 [Probe](https://github.com/SamYuan1990/Probe) 项目，该项目的目的就是测试 Fabric 的最佳出块参数。
