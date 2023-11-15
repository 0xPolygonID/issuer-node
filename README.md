# Polygon ID Issuer Node

[![Checks](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/checks.yml/badge.svg)](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/checks.yml)
[![golangci-lint](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/golangci-lint.yml)

TODO: Add a better project description

This is a set of tools and APIs for issuers of zk-proof credentials, designed to be extensible. It allows an authenticated user to create schemas for issuing and managing credentials of identities. It also provides a [user interface](ui/README.md) to manage issuer schemas, credentials, issuer state and connections.

This repository is for anyone to create their own [issuer node](https://0xpolygonid.github.io/tutorials/issuer-node/issuer-node-overview/) for Polygon ID.

---

## Table of Contents

- [Installation](#installation)
    - [Prerequisites](#Prerequisites)
    - [Issuer Node](#Installation-Issuer-Node)
    - [UI](#Installation-UI)
- [Quick Start Demo](#quick-start-demo)
- [Documentation](#documentation)
- [Tools](#tools)
- [Contributing](#contributing)
- [License](#license)

## Installation
> [!INFO]
> The provided installation guide is non-production ready, in order to accomplish it, please visit this reference one TODO insert reference to standalone.

### Prerequisites

- Unix-based operating system (e.g. Debian, Arch, Mac OS)
- [Docker Engine](https://docs.docker.com/engine/) `1.27+`
- Makefile toolchain `GNU Make 3.81`
- Publicly accessible URL - For testing purposes only you can use any of this
    - [Ngrok](https://ngrok.com/)
    - [Localtunnel](https://theboroer.github.io/localtunnel-www/)
- Polygon Mumbai or Main RPC - You can get one in any of the providers of this list
    - [Chainstack](https://chainstack.com/)
    - [Ankr](https://ankr.com/)
    - [QuickNode](https://quicknode.com/)
    - [Alchemy](https://www.alchemy.com/)
    - [Infura](https://www.infura.io/)


> [!INFO]
> There is no compatibility with Windows environments at this time.

To help expedite a lot of the Docker commands, many have been abstracted using `make` commands. Included in the following sections are the equivalent Docker commands that show what is being run.


### Installation Issuer Node

### Installation UI


## Quick Start Demo

Verify the correct installation of Issuer Node, **issue** and **verify** your first **verifiable credential** with this [Quick start demo](https://devs.polygonid.com/docs/quick-start-demo/)!

## Documentation

* [Issuer Node resources](https://devs.polygonid.com/docs/category/issuer/)
* [Polygon ID core concepts](https://devs.polygonid.com/docs/introduction/)

## Tools
> [!WARNING]
> **Demo Issuer** and **Verifier Demo** are for **testing** purposes **only**.


* [Schema Builder](https://schema-builder.polygonid.me/) - Create your custom schemas.
* [Demo Issuer UI](https://user-ui:password-ui@issuer-ui.polygonid.me/)
* [Verifier Demo](https://verifier-demo.polygonid.me/)

## License

See [LICENSE](LICENSE.md).