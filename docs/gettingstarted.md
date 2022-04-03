# Quick start

## recording demo

[here](https://www.bilibili.com/video/BV1k5411L79A/) is a video demo。

## install

You are able to install `tape` via ways blow：

1. **down load binary**：from[release page](https://github.com/hyperledger-twgc/tape/releases)download binary。

2. **local complie**：
```shell
make tape
```

3. **Docker image**: 
```shell
docker pull ghcr.io/hyperledger-twgc/tape
```

## local Docker complie(optional)
```shell
make docker
```

## edit config file apply with your fabric network

according to [config file](configfile.md) edit config file.

## run
### default run：
```shell
./tape --config=config.yaml --number=40000
```
### CommitOnly
```shell
docker run -v $PWD:/tmp ghcr.io/hyperledger-twgc/tape tape commitOnly -c $CONFIG_FILE -n 40000
```
### EndorsementOnly
```shell
docker run -v $PWD:/tmp ghcr.io/hyperledger-twgc/tape tape endorsementOnly -c $CONFIG_FILE -n 40000
```

## log level
environment `TAPE_LOGLEVEL` for log level
```shell
export TAPE_LOGLEVEL=debug
```

default is `warn`：
- panic
- fatal
- error
- warn
- warning
- info
- debug
- trace
