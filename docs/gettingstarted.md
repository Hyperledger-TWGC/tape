# Quick start

## Tutorial Video

[here](https://www.bilibili.com/video/BV1k5411L79A/) is a video demo。

## Install

You are able to install `tape` via ways blow：

1. **download binary**：from[release page](https://github.com/hyperledger-twgc/tape/releases)download binary。

2. **local complie**：
```shell
make tape
```

3. **Docker image**: 
```shell
docker pull ghcr.io/hyperledger-twgc/tape
```

## Local Docker complie(optional)
```shell
make docker
```

## Edit configuration file according to your fabric network

according to [this](configfile.md) edit config file.

## Run
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
### Prometheus
```shell
./tape --config=config.yaml --number=40000 --prometheus
```
and the Prometheus will listen `:8080/metrics`, the metrics names as `transaction_latency_duration` and `read_latency_duration` for now, they are float based time duration.

## Log level
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
