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
    - [Run Issuer Node API and UI (docker compose with images from privadoid dockerhub registry)](#run-issuer-node-api-and-ui-docker-compose-with-images-from-privadoid-registry)
    - [Install and run Issuer Node API and UI (docker compose and build from source)](#install-and-run-issuer-node-api-and-ui---docker-compose-and-build-from-source)
    - [Running only Issuer Node API (docker compose and build from source)](#running-only-issuer-node-api-docker-compose-and-build-from-source)
  - [KMS Providers Configuration](#kms-providers-configuration)
    - [Vault](#Running-issuer-node-with-vault-instead-of-local-storage-file)
    - [AWS Secret Manager](#Running-issuer-node-with-AWS-Secret-Manager)
    - [AWS KMS](#Running-issuer-node-with-AWS-KMS)
  - [Quick Start Demo](#quick-start-demo)
  - [Documentation](#documentation)
  - [Tools](#tools)
  - [License](#license)

## Quick Start Installation
> [!NOTE]
> The provided installation guide is **non-production** ready. For production deployments please refer to  [Standalone Mode Guide](https://docs.privado.id/docs/issuer/setup-issuer-core).
>
> There is no compatibility with Windows environments at this time. While using WSL should be ok, it's not officially supported.
> 
> **After changing the configuration, you must restart the issuer node docker containers.**

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

1. Copy the config sample files:
```shell
cp .env-issuer.sample .env-issuer
cp .env-ui.sample .env-ui
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
define the RPCs that we will use.

4. Run
```shell
make run-all-registry
```

after a few seconds, the issuer node will be running and you can check the docker containers with `docker ps` and you 
should see something like this:
```shell
CONTAINER ID   IMAGE                           COMMAND                  CREATED          STATUS                    PORTS                                        NAMES
6e923fa11228   privadoid/issuernode-ui         "/bin/sh /app/script…"   37 seconds ago   Up 32 seconds (healthy)   0.0.0.0:8088->80/tcp                         issuer-ui-1
16afc9d66591   privadoid/issuernode-api        "sh -c ./pending_pub…"   37 seconds ago   Up 32 seconds (healthy)                                                issuer-pending_publisher-1
ceb41877c041   privadoid/issuernode-api        "sh -c ./notificatio…"   37 seconds ago   Up 32 seconds (healthy)                                                issuer-notifications-1
bd7b69984f1c   privadoid/issuernode-api        "sh -c './migrate &&…"   38 seconds ago   Up 34 seconds (healthy)   0.0.0.0:3001->3001/tcp                       issuer-api-1
25ae0fcac183   postgres:14-alpine              "docker-entrypoint.s…"   38 seconds ago   Up 36 seconds (healthy)   5432/tcp                                     issuer-postgres-1
a4a1d3ec9159   redis:6-alpine                  "docker-entrypoint.s…"   38 seconds ago   Up 36 seconds (healthy)   6379/tcp                                     issuer-redis-1
```


5. Import your Ethereum private key

```shell
make private_key=<private-key> import-private-key-to-kms 
```

**_TODO_**: Add section to configure payments

then visit:
* http://localhost:8088/ to access the UI (default username / password are: user-ui, password-ui). You can set them using env [vars](.env-ui.sample).
* <PUBLICLY_ACCESSIBLE_URL_POINTING_TO_ISSUER_SERVER_PORT>:3001/ to access the API. (default username / password are: user-issuer, password-issuer)
  You can set them using env [vars](.env-issuer.sample).

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
# API Auth credentials - You can change these values
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

**_TODO_**: Add section to configure payments****

4. Copy .env-ui sample file and fill the needed env variables:

```bash 
cp .env-ui.sample .env-ui
```
If you want to disable UI authentication just change `ISSUER_UI_INSECURE=true`, or if you want to change ui authentication:

*.env-ui*
```bash
ISSUER_UI_AUTH_USERNAME=<your-username>
ISSUER_UI_AUTH_PASSWORD=<your-password>
```

5. Run API, UI and infrastructure (Postgres, localstorage and Redis)

To do a build and start both the API and the UI in a single step, you can use the following command:
```bash
make run-all
```
then visit 
* http://localhost:8088/ to access the UI
* <PUBLICLY_ACCESSIBLE_URL_POINTING_TO_ISSUER_SERVER_PORT>:3001/ to access the API.

6. Import your Ethereum private Key:
Configure the private key. This step is needed in order to be able to transit the issuer's state. To perform that
action the given account has to be funded. For Amoy network you can request some testing Matic [here](https://www.alchemy.com/faucets/polygon-amoy)
```bash
make private_key=<private-key> import-private-key-to-kms
```


### Running only Issuer Node API (docker compose and build from source)

If you want to run only the API, you can follow the steps below. You have to have the .env-issuer file filled with 
the proper values and the `resolvers_settings.yaml` file with the proper RPCs.
Make sure the infrastructure is running (Postgres, localstorage and Redis). If not, you can run it with the following command:

```shell
make up
```

Then run: 

```shell
make build-api && make run-api
```
----
**Troubleshooting:**

In order to **stop all** the containers, run the following command:

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
Consider that if you have the issuer node running, after changing the configuration you must restart all the containers.
In all options the **.env-issuer** file is necessary.

#### Running issuer node with vault instead of local storage file
The issuer node can be configured to use a [HashiCorp Vault](https://www.vaultproject.io), as kms provider.
However, Vault needs a [plugin](https://github.com/iden3/vault-plugin-secrets-iden3) 
for key generation and message signing. This is because the issuer node does not generate private keys, but rather 
delegates that action and the signing of messages to the vault.

Setup environment variables in `.env-issuer` file:

```bash
ISSUER_KMS_BJJ_PROVIDER=vault
ISSUER_KMS_ETH_PROVIDER=vault
```
After configuring the variables, run the following commands:

```bash
make up
``` 
In this case, the docker container for vault will be created.

To import the ethereum private key (if you have changed the kms provider you have to import the private key again) necessary to 
transition issuer node states onchain, the command is the same as explained before:

```shell
make private_key <private-key> import-private-key-to-kms
```
You should get something like this:

```shell
 ... private key saved to vault: path:=pbkey
```

#### Running issuer node with AWS Secret Manager
Another alternative is to configure the issuer node to store the private keys of the identities in the AWS Secret Manager service. 
Both babyjubjub type keys and ethereum keys can be stored using this service. To configure the issuer node, you must 
change the following variables in the .env-issuer file:

```shell
ISSUER_KMS_BJJ_PROVIDER=aws-sm
ISSUER_KMS_ETH_PROVIDER=aws-sm
ISSUER_KMS_AWS_ACCESS_KEY=<your-aws-access-key>
ISSUER_KMS_AWS_SECRET_KEY=<your-aws-secret-key>
ISSUER_KMS_AWS_REGION=<your-aws-region>
```

After configuring the variables, run the following commands:
```shell
make up
```

Then you must run the command to import the ethereum private key to the kms.:

```shell
make private_key <private-key> import-private-key-to-kms`
 ```
If all went well, you should see something like this:
```shell
 ... private key saved to aws: path:=pbkey
```

#### Running issuer node with AWS KMS
Another alternative is to configure the issuer node to store the private keys of the identities in the AWS KMS service.
**Only ethereum keys** can be stored using this service. To configure the issuer node, you must change the following variables in the .env-issuer file:

```shell

```shell
ISSUER_KMS_BJJ_PROVIDER= [localstorage | vault | aws-sm] 
ISSUER_KMS_ETH_PROVIDER=aws-kms
ISSUER_KMS_AWS_ACCESS_KEY=<your-aws-access-key>
ISSUER_KMS_AWS_SECRET_KEY=<your-aws-secret-key>
ISSUER_KMS_AWS_REGION=<your-aws-region>
```
After configuring the variables, run the following commands:
```shell
make up
```
then you must run the command `make private_key <private-key> import-private-key-to-kms` to import the ethereum private key to the kms.
If all went well, you should see something like this:

```bash
 ... Key material successfully imported!!!
```

## Quick Start Demo

This [Quick Start Demo](https://docs.privado.id/docs/quick-start-demo) will walk you through the process of **issuing** and **verifying** your **first credential**.

## Documentation

* [Issuer Node resources](https://docs.privado.id/docs/category/issuer)
* [Privado ID core concepts](https://docs.privado.id/docs/introduction)

## Tools
> [!WARNING]
> **Privado ID Demo Issuer** and **Privado ID Verifier Backend Demo** are for **testing** purposes **only**.

* [Privado ID Schema Explorer](https://tools.privado.id): Browse and reuse community-maintained VC schemas.
* [Privado ID Schema Builder](https://tools.privado.id/builder): A tool for easy VC schema creation.
* [Privado ID Demo Issuer](https://user-ui:password-ui@issuer-node-testing.privado.id): A demo version of the Issuer Node in action.
* [Privado ID Query Builder](https://tools.privado.id/query-builder): Build and customize queries to verify VCs.
* [Privado ID Verifier Backend Demo](https://verifier-backend.privado.id): Demo of our [reference implementation](https://github.com/0xPolygonID/verifier-backend) of a verifier backend.
* [Privado ID Web Wallet](https://wallet.privado.id): Identity and credential wallet for web clients.
* [Privado ID Mobile Wallet for Android](https://play.google.com/store/apps/details?id=id.privado.wallet): Wallet for Android devices.
* [Privado ID Mobile Wallet for iOS](https://apps.apple.com/us/app/privadoid/id6566184703): Wallet for iOS devices.
* [Privado ID Marketplace](https://marketplace.privado.id): Explore credentials submitted by trusted issuers.

## License

By contributing to this project, you agree to the terms of licenses [Apache](LICENSE-APACHE) and [Mit](LICENSE-MIT).
