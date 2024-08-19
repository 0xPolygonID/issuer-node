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
    - [Issuer Node API](#issuer-node-api)
      - [Deploy Issuer Node Infrastructure](#deploy-issuer-node-infrastructure)
      - [Run Issuer Node API](#run-issuer-node-api)
    - [Issuer Node UI](#issuer-node-ui)
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

### Issuer Node API

In this section we will cover the installation of the Issuer Node API.

> [!NOTE]
> This Quick Installation Guide is prepared for Polygon Amoy (Testnet) both for the state contract and issuer dids. If you want to deploy the node with Polygon Main configuration, please visit our [advanced Issuer Node configuration guide](https://devs.polygonid.com/docs/issuer/issuer-configuration/)).


#### Deploy Issuer Node Infrastructure

1. Copy the config sample files:

    ```shell
    cp .env-issuer.sample .env-issuer
    ```

2. Copy the resolvers settings sample file:

    ```shell
    cp resolvers_settings_sample.yaml resolvers_settings.yaml
    ```
3. Fill the `networkURL` in the resolvers_settings.yaml file with the proper value for every chain:

 
3. Start the infrastructure:

    ```bash
    make up
    ```

4. Enable vault authentication:

    ```bash
    make add-vault-token
    ```

5. Write the private key in the vault. This step is needed in order to be able to transit the issuer's state. To perform that action the given account has to be funded. For Amoy network you can request some testing Matic [here](https://www.alchemy.com/faucets/polygon-amoy).

    ```bash
    make private_key=<YOUR_WALLET_PRIVATE_KEY> import-private-key-to-kms
    ```

----
**Troubleshooting:**

In order to **stop** and **delete** all the containers.

> [!WARNING]
> This will permanently delete all data.

``` bash
make down
```

If you experience **problems** with the **vault**, follow these commands:

``` bash
docker stop issuer-vault-1    // Stops the container issuer-vault-1 
docker rm issuer-vault-1      // Removes container issuer-vault-1
make clean-vault              // Removes all the data in the vault, including the token
make up                       // Starts the database, cache and vault storage (i.e, postgres, redis and vault)
```
Wait 20 secs so the vault can boot and generate a token.

``` bash
make add-vault-token                                          // Adds the generated token to the ISSUER_KEY_STORE_TOKEN var in .env-issuer
make private_key=<YOUR_WALLET_PRIVATE_KEY> add-private-key    // Stores the private key in the vault
```

----
#### Run Issuer Node API

The issuer node is extensively configurable, for a detailed list of the configuration, please visit our [detailed configuration guide](https://devs.polygonid.com/docs/issuer/issuer-configuration/).

1. Fill the .env-issuer config file with the proper variables:

*.env-issuer*
```bash
ISSUER_API_AUTH_USER=user-issuer
ISSUER_API_AUTH_PASSWORD=password-issuer
ISSUER_SERVER_URL=<PUBLICLY_ACCESSIBLE_URL_POINTING_TO_ISSUER_SERVER_PORT>
```

2. Run api:

```bash
make build # Only needed the first time
make run
```

> Core API specification - http://localhost:3001/

---

**Troubleshooting:**
---

### Issuer Node UI

In this section we will cover the installation of the Issuer Node UI, before continuing with these steps, make sure that you have followed the [Deploy Issuer Node Infrastructure](#Deploy-Issuer-Node-Infrastructure) section before continuing.

In order to make the UI work, we will need configure some env variables in the `.env-api` file

1. Copy .env-ui sample file and fill the needed env variables:

```bash 
cp .env-ui.sample .env-ui
```

If you want to use basic auth, you can set the following variables in the `.env-ui` file:
```bash
ISSUER_UI_INSECURE=false
ISSUER_UI_AUTH_USERNAME=user-ui
ISSUER_UI_AUTH_PASSWORD=password-ui
```


3. Run UI:
This command will start the UI in the port 8088 and the API in the port 3001.
```bash
make run-ui
```
then you can access the UI in the following URL:
>**UI** - http://localhost:8088/


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