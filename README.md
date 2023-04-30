# Tape
<div align="center">
<img src="logo.svg" width="100">
</div>

[![Go doc](https://img.shields.io/badge/go.dev-reference-brightgreen?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/hyperledger-twgc/tape)
[![Github workflow test](https://github.com/Hyperledger-TWGC/tape/actions/workflows/test.yml/badge.svg)](https://github.com/Hyperledger-TWGC/tape/actions/workflows/test.yml)

A light-weight tool to test performance of Hyperledger Fabric

It is used to perform super simple performance test.
Our main focus is to make sure that *tape will not be the bottleneck of performance test*

README in English/[中文](README-zh.md)


## Table Of Content

- [Tape](#tape)
  - [Table Of Content](#table-of-content)
  - [Prerequisites](#prerequisites)
  - [Configure](#configure)
  - [Usage](#usage)
    - [Binary](#binary)
    - [Docker](#docker)
  - [FAQ](#faq)
  - [Contribute](#contribute)
  - [License](#license)
  - [Contact](#contact)
  - [Credits](#credits)
  - [Contributors](#contributors)

---
## Prerequisites

You could get `tape` in three ways:
1. Download binary: get release tar from [release page](https://github.com/hyperledger-twgc/tape/releases), and extract `tape` binary from it
2. Build from source: clone this repo and run `make tape` at root dir. Go1.14 or higher is required. `tape` binary will be available at project root directory.
3. Pull docker image: `docker pull ghcr.io/hyperledger-twgc/tape`
---

## [Configure](docs/configfile.md)

## Usage
For further usage, please see [getting start](/docs/gettingstarted.md)

### Binary
Execute `./tape -c config.yaml -n 40000` to generate 40000 transactions to Fabric.


### Docker
```shell
docker run -v $PWD:/tmp ghcr.io/hyperledger-twgc/tape tape -c $CONFIG_FILE -n 40000
```

*Set this to integer times of batchsize, so that last block is not cut due to timeout*. For example, if you have batch size of 500, set this to 500, 1000, 40000, 100000, etc.

## [FAQ](https://github.com/Hyperledger-TWGC/tape/wiki/FAQ)

---
## Contribute
[How to Contribute](CONTRIBUTING.md)
[workflow](docs/workflow.md)

---
## License
Hyperledger Project source code files are made available under the Apache License, Version 2.0 (Apache-2.0), located in the [LICENSE](LICENSE) file.

---
## Contact

* [Maintainers](MAINTAINERS.md)
---

## Credits

Icons made by <a href="https://www.flaticon.com/authors/good-ware" title="Good Ware">Good Ware</a> from <a href="https://www.flaticon.com/" title="Flaticon">www.flaticon.com</a>

---
## Contributors

<a href="https://github.com/Hyperledger-TWGC/tape/graphs/contributors">
  <img src="https://contributors-img.web.app/image?repo=Hyperledger-TWGC/tape" />
</a>