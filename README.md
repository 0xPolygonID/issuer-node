# Privado ID Issuer Node

[![Checks](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/checks.yml/badge.svg)](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/checks.yml)
[![golangci-lint](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/golangci-lint.yml)

Streamline the **Verifiable Credentials issuance** process with the user-friendly API and UI of the Issuer Node within the Privado ID ecosystem. The on-premise (self-hosted) Issuer Node, seamlessly integrated with a robust suite of tools including the mobile Wallet, Schema Builder, and Credential Marketplace, guarantees a frictionless experience for effortlessly issuing and verifying credentials.

![Triagle-of-trust](docs/assets/img/triangle-of-trust.png)

**Features:**

* Create Issuer Identities.
* Issue VCs.
* Revoke VCs.
* Fetch VCs.
* Transit Issuer's state.
* Create Issuer-User connections.
* Issuer's UI.

---

## Table of Contents
- [Privado ID Issuer Node](#privado-id-issuer-node)
  - [Table of Contents](#table-of-contents)
  - [Quick Start Installation](#quick-start-installation)
    - [Prerequisites](#prerequisites)
    - [Install and run Issuer Node API and UI](#install-and-run-issuer-node-api-and-ui)
    - [Running only Issuer Node API](#running-only-issuer-node-api)
  - [KMS Providers Configuration](#kms-providers-configuration)
  - [Quick Start Demo](#quick-start-demo)
  - [Documentation](#documentation)
  - [Tools](#tools)
  - [License](#license)

## Quick Start Installation
> [!NOTE]
> The provided installation guide is **non-production** ready. For production deployments please refer to  [Standalone Mode Guide](https://devs.polygonid.com/docs/issuer/setup-issuer-core/).
>
> There is no compatibility with Windows environments at this time. While using WSL should be ok, it's not officially supported.

### Prerequisites

- Unix-based operating system (e.g. Debian, Arch, Mac OS)
- [Docker Engine](https://docs.docker.com/engine/) `1.27+`
- Makefile toolchain `GNU Make 3.81`
- Publicly accessible URL - The issuer node API must be publicly reachable. Please make sure you properly configure your proxy or use a tool like [Localtunnel](https://theboroer.github.io/localtunnel-www/) for testing purposes.
- Polygon Amoy or Main RPC - You can get one in any of the providers of this list
    - [Chainstack](https://chainstack.com/)
    - [Ankr](https://ankr.com/)
    - [QuickNode](https://quicknode.com/)
    - [Alchemy](https://www.alchemy.com/)
    - [Infura](https://www.infura.io/)

### Run Issuer Node API and UI (docker compose with images from privadoid registry)
To run the issuer node (API and UI) quickly and without too many customizations follow the following steps:
1. Copy the config sample file:
```shell
cp .env-issuer.sample .env-issuer
```
2. Fill the .env-issuer config file with the proper variables:

*.env-issuer*
```bash
ISSUER_SERVER_URL=<PUBLICLY_ACCESSIBLE_URL_POINTING_TO_ISSUER_SERVER_PORT>
```

3. Create a file with the networks' configuration. You can copy and modify the provided sample file:

```bash
cp resolvers_settings_sample.yaml resolvers_settings.yaml
```
then modify the file with the proper values. The most important fields to run the issuer node are RPC (`networkURL`) fields.
In this file you can define customizations for each type of blockchain and network. For this example, we only need to
define the RPCs. that will use.

4. Run
```shell
make run-full
```

after a few seconds, the issuer node will be running and you can check the docker containers with `docker ps` and you 
should see something like this:
```shell
CONTAINER ID   IMAGE                           COMMAND                  CREATED          STATUS                    PORTS                                        NAMES
6e923fa11228   privadoid/issuernode-ui         "/bin/sh /app/script…"   37 seconds ago   Up 32 seconds (healthy)   0.0.0.0:8088->80/tcp                         issuer-ui-1
16afc9d66591   privadoid/issuernode            "sh -c ./pending_pub…"   37 seconds ago   Up 32 seconds (healthy)                                                issuer-pending_publisher-1
ceb41877c041   privadoid/issuernode            "sh -c ./notificatio…"   37 seconds ago   Up 32 seconds (healthy)                                                issuer-notifications-1
bd7b69984f1c   privadoid/issuernode            "sh -c './migrate &&…"   38 seconds ago   Up 34 seconds (healthy)   0.0.0.0:3001->3001/tcp                       issuer-api-1
25ae0fcac183   postgres:14-alpine              "docker-entrypoint.s…"   38 seconds ago   Up 36 seconds (healthy)   5432/tcp                                     issuer-postgres-1
a4a1d3ec9159   redis:6-alpine                  "docker-entrypoint.s…"   38 seconds ago   Up 36 seconds (healthy)   6379/tcp                                     issuer-redis-1
```


5. Import your metamask private key

```shell
make private_key=<private-key> import-private-key-to-kms 
```

then visit:
* http://localhost:8088/ to access the UI
* http://localhost:3001/ to access the API. (**user: user-issuer**, **password: password-issuer**)

**Different installation alternatives can be seen later.**

### Install and run Issuer Node API and UI - docker compose and build from source.
> [!NOTE]
> This Quick Installation Guide is prepared for Polygon Amoy (Testnet) both for the state contract and issuer dids.

In this section we will see how to install the issuer node api and the UI along with the necessary infrastructure in 
the most basic way, without too much customization.

1. Copy the config sample file:
```shell
cp .env-issuer.sample .env-issuer
```

2. Fill the .env-issuer config file with the proper variables:

*.env-issuer*
```bash
ISSUER_SERVER_URL=<PUBLICLY_ACCESSIBLE_URL_POINTING_TO_ISSUER_SERVER_PORT>
# API Auth
ISSUER_API_AUTH_USER=user-issuer
ISSUER_API_AUTH_PASSWORD=password-issuer
```
3. Create a file with the networks' configuration. You can copy and modify the provided sample file:

```bash
cp resolvers_settings_sample.yaml resolvers_settings.yaml
```
then modify the file with the proper values. The most important fields to run the issuer node are RPC (`networkURL`) fields.
In this file you can define customizations for each type of blockchain and network. For this example, we only need to 
define the RPCs. that will use.

4. Copy .env-ui sample file and fill the needed env variables:

```bash 
cp .env-ui.sample .env-ui
```
If you want to enable UI authentication, you must change the value of the following variables:

*.env-ui*
```bash
ISSUER_UI_INSECURE=false
ISSUER_UI_AUTH_USERNAME=user-ui
ISSUER_UI_AUTH_PASSWORD=password-ui
```

5. Run API, UI and infrastructure (Postgres, localstorage and Redis)

To do a build and start both the API and the UI in a single step, you can use the following command:
```bash
make run-all
```
then visit 
* http://localhost:8088/ to access the UI
* http://localhost:3001/ to access the API.

6. Import your metamask private Key:
Write the private. This step is needed in order to be able to transit the issuer's state. To perform that
   action the given account has to be funded. For Amoy network you can request some testing Matic [here](https://www.alchemy.com/faucets/polygon-amoy)
```bash
make private_key=<private-key> import-private-key-to-kms
```


### Running only Issuer Node API

If you want to run only the API, you can follow the steps below. You have to have the .env-issuer file filled with 
the proper values and the resolver_settings.yaml file with the proper RPCs.
Then run: 

```shell
make build && make run
```
----
**Troubleshooting:**

In order to **stop** **all** the containers, run the following command:

> [!NOTE] This will not delete the data in the vault and the database.

``` bash
make stop-all
```

To stop only the API and UI container, run:

``` bash
make stop
```

If you want to **delete** all the data in the vault and the database, run:

``` bash
make clean-volumes
```

If for some reason you only need to restart the UI, run:

``` bash
make run-ui
```

To restart the api after changes (pull code with changes):

```bash 
make build && make run
```

### KMS Providers Configuration
Consider that if you have the issuer node running, after changing the configuration you must restart it.
In all options the .env-issuer file is necessary.

#### Running issuer node with vault instead of local storage file
The issuer node can be configured to use a local storage, that is, a local file, as kms provider. 
This alternative can be useful in development or testing environments. To do it:

Setup environment variables in `.env-issuer` file:

```bash
ISSUER_KMS_BJJ_PROVIDER=vault
ISSUER_KMS_ETH_PROVIDER=vault
```

To import the private key necessary to transition onchain states, the command is the same as [explained before](#install-and-run-issuer-node-api-and-ui).

## Quick Start Demo

This [Quick Start Demo](https://devs.polygonid.com/docs/quick-start-demo/) will walk you through the process of **issuing** and **verifying** your **first credential**.

## Documentation

* [Issuer Node resources](https://devs.polygonid.com/docs/category/issuer/)
* [Privado ID core concepts](https://devs.polygonid.com/docs/introduction/)

## Tools
> [!WARNING]
> **Demo Issuer** and **Verifier Demo** are for **testing** purposes **only**.


* [Schema Builder](https://schema-builder.polygonid.me/) - Create your custom schemas to issue VC.
* [Demo Issuer UI](https://user-ui:password-ui@issuer-ui.polygonid.me/) - Test our Issuer Node UI.
* [Verifier Demo](https://verifier-demo.polygonid.me/) - Verify your VCs.
* [Polygon ID Android Mobile App](https://play.google.com/store/apps/details?id=com.polygonid.wallet&hl=en&gl=US)
* [Polygon ID IOS Mobile App](https://apps.apple.com/us/app/polygon-id/id1629870183)
* [Marketplace](https://marketplace.polygonid.me/) - Explore credentials submitted by trusted issuers.

## License

See [LICENSE](LICENSE.md).