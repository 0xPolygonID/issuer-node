# Polygon ID Issuer Node

[![Checks](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/checks.yml/badge.svg)](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/checks.yml)
[![golangci-lint](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/golangci-lint.yml)

Streamline the **Verifiable Credentials issuance** process with the user-friendly API and UI of the Issuer Node within the Polygon ID ecosystem. The on-premise (self-hosted) Issuer Node, seamlessly integrated with a robust suite of tools including the mobile Wallet, Schema Builder, and Credential Marketplace, guarantees a frictionless experience for effortlessly issuing and verifying credentials.

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

- [Quick Start Installation](#quick-start-installation)
    - [Prerequisites](#Prerequisites)
    - [Issuer Node Api](#issuer-node-api)
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
- Polygon Mumbai or Main RPC - You can get one in any of the providers of this list
    - [Chainstack](https://chainstack.com/)
    - [Ankr](https://ankr.com/)
    - [QuickNode](https://quicknode.com/)
    - [Alchemy](https://www.alchemy.com/)
    - [Infura](https://www.infura.io/)

### Issuer Node API

In this section we will cover the installation of the Issuer Node API.

> [!NOTE]
> This Quick Installation Guide is prepared for Polygon Mumbai (Testnet) both for the state contract and issuer dids. If you want to deploy the node with Polygon Main configuration, please visit our [advanced Issuer Node configuration guide](https://devs.polygonid.com/docs/issuer/issuer-configuration/)).


#### Deploy Issuer Node Infrastructure

1. Copy the config sample files:

    ```shell
    cp .env-issuer.sample .env-issuer
    cp .env-api.sample .env-api
    ```

2. Fill the .env-issuer config file with the proper variables:

   *.env-issuer*
    ```shell
    ISSUER_ETHEREUM_URL=<YOUR_RPC_PROVIDER_URI_ENDPOINT>
    ```
3. Start the infrastructure:

    ```shell
    make up
    ```

4. Enable vault authentication:

    ```bash
    make add-vault-token
    ```

5. Write the private key in the vault. This step is needed in order to be able to transit the issuer's state. To perform that action the given account has to be funded. For Mumbai network you can request some testing Matic [here](https://mumbaifaucet.com/).

    ```bash
    make private_key=<YOUR_WALLET_PRIVATE_KEY> add-private-key
    ```

----
**Troubleshooting:**

In order to **stop** and **delete** all the containers.

> [!WARNING]
> This will permanently delete all data, making it necessary to create an Issuer DID again.

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
    make run
    ```

> Core API specification - http://localhost:3001/

---

**Troubleshooting:**

Restart the api.

```bash 
make restart-api
```

---

### Issuer Node UI

In this section we will cover the installation of the Issuer Node UI, before continuing with these steps, make sure that you have followed the [Deploy Issuer Node Infrastructure](#Deploy-Issuer-Node-Infrastructure) section before continuing.

In order to make the UI work, we will need configure some env variables in the `.env-api` file

1. Copy .env-ui sample file and fill the needed env variables:


    ```bash 
    cp .env-ui.sample .env-ui
    ```

    *.env-ui*
    ```bash
    ISSUER_UI_AUTH_USERNAME=user-ui
    ISSUER_UI_AUTH_PASSWORD=password-ui
    ```
    
    *.env-api*
    ```bash
    ISSUER_API_UI_SERVER_URL={PUBLICLY_ACCESSIBLE_URL_POINTING_TO_ISSUER_API_UI_SERVER_PORT}
    ```

    > **_NOTE:_**  It is possible to register custom did methods. This field accepts an array of objects in JSON format.</br>
    > Example:
        ```
        ISSUER_CUSTOM_DID_METHODS='[{"blockchain":"linea","network":"testnet","networkFlag":"0b01000001","chainID":59140}]'
        ```

2. Generate Issuer DID:

    ```bash
    make generate-issuer-did
    ```

3. Run UI:

    ```bash
    make run-ui
    ```


>**API UI specification** - http://localhost:3002/
> 
>**UI** - http://localhost:8088/

---
**Troubleshooting:**

Restart the ui:

```bash 
make restart-ui
```
---

## Quick Start Demo

This [Quick Start Demo](https://devs.polygonid.com/docs/quick-start-demo/) will walk you through the process of **issuing** and **verifying** your **first credential**.

## Documentation

* [Issuer Node resources](https://devs.polygonid.com/docs/category/issuer/)
* [Polygon ID core concepts](https://devs.polygonid.com/docs/introduction/)

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