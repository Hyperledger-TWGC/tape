# Tape
A light-weight tool to test performance of Hyperledger Fabric

English/[中文](README-zh.md)

[![Build Status](https://dev.azure.com/guojiannan1101/guojiannan1101/_apis/build/status/guoger.tape?branchName=master)](https://dev.azure.com/guojiannan1101/guojiannan1101/_build/latest?definitionId=1&branchName=master)

<img src="logo.svg" width="100">

## Why Tape

Sometimes we need to test performance of a deployed Fabric network with ease. There are many excellent projects out there, i.e. Hyperledger Caliper. However, we sometimes just need a tiny, handy tool, like `tape`.

## What is it
This includes a very simple traffic generator:
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

Our main focus is to make sure that *tape will not be the bottleneck of performance test*

## Usage

### Install

You could get `tape` in three ways:
1. Download binary: get release tar from [release page](https://github.com/guoger/tape/releases), and extract `tape` binary from it
2. Build from source: clone this repo and run `make tape` at root dir. Go1.14 or higher is required. `tape` binary will be available at project root directory.
3. Pull docker image: `docker pull guoger/tape`

### [Configure](docs/configfile.md)

### Run

#### Binary

Execute `./tape run -c config.yaml -n 40000` to generate 40000 transactions to Fabric.

#### Docker

```
docker run -v $PWD:/tmp guoger/tape tape -c $CONFIG_FILE -n 40000
```

*Set this to integer times of batchsize, so that last block is not cut due to timeout*. For example, if you have batch size of 500, set this to 500, 1000, 40000, 100000, etc.

## Tips

- We use [logrus](https://github.com/sirupsen/logrus) for logging, which can be set with env var `export TAPE_LOGLEVEL=debug`.
Here are possbile values (warn by default)
`"panic", "fatal", "error", "warn", "warning", "info", "debug", "trace"`

- Put this generator closer to Fabric, or even on the same machine. This is to prevent network bandwidth from being the bottleneck.

- Increase number of messages per block in your channel configuration may help
- [Workflow](docs/workflow.md)



## [How to Contribute](CONTRIBUTING.md)

## [Maintainers](MAINTAINERS.md)


## LICENSE

Hyperledger Project source code files are made available under the Apache License, Version 2.0 (Apache-2.0), located in the [LICENSE](LICENSE) file.

## Credits

Icons made by <a href="https://www.flaticon.com/authors/good-ware" title="Good Ware">Good Ware</a> from <a href="https://www.flaticon.com/" title="Flaticon">www.flaticon.com</a>
