# Tape : Aplha

## A light-weight tool to test performance of Hyperledger Fabric

English/[中文](README-zh.md)

---
[**Sample run of Tape**](https://www.bilibili.com/video/BV1k5411L79)

---
## Table Of Content

* [Prerequisites](#prerequisites)
* [Configure](docs/configfile.md)
* [Usage](#usage)
* [Contributing](#contributing)
* [License](#license)
* [Contact](#contact)
* [Regards](#thanks-for-choosing)

---
## Prerequisites

You could get `tape` in three ways:
1. Download binary: get release tar from [release page](https://github.com/hyperledger-twgc/tape/releases), and extract `tape` binary from it
2. Build from source: clone this repo and run `make tape` at root dir. Go1.14 or higher is required. `tape` binary will be available at project root directory.
3. Pull docker image: `docker pull ghcr.io/hyperledger-twgc/tape`

3. Pull docker image: `docker pull guoger/tape` or `docker pull 19902439/tapealpha`(for alpha only)
---

## [Configure](docs/configfile.md)

## Usage

### Binary

Execute `./tape -c config.yaml -n 40000` to generate 40000 transactions to Fabric.


### Docker

```
docker run -v $PWD:/tmp ghcr.io/hyperledger-twgc/tape tape -c $CONFIG_FILE -n 40000
```

*Set this to integer times of batchsize, so that last block is not cut due to timeout*. For example, if you have batch size of 500, set this to 500, 1000, 40000, 100000, etc.

### CommitOnly
```

docker run -v $PWD:/tmp guoger/tape tape commitOnly -c $CONFIG_FILE -n 40000

```


### EndorsementOnly
```

docker run -v $PWD:/tmp guoger/tape tape endorsementOnly -c $CONFIG_FILE -n 40000

```

---
## Contributing
Contributions are what make the open source community such an amazing place to be learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request
6. [How to Contribute](CONTRIBUTING.md)

---
## License
Hyperledger Project source code files are made available under the Apache License, Version 2.0 (Apache-2.0), located in the [LICENSE](LICENSE) file.

---
## Contact

* [Maintainers](MAINTAINERS.md)
---

### THANKS FOR CHOOSING

