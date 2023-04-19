# Polygon ID Issuer Node

[![Checks](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/checks.yml/badge.svg)](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/checks.yml)
[![golangci-lint](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/golangci-lint.yml)

This is a set of tools and APIs for issuers of zk-proof credentials, designed to be extensible. It allows an authenticated user to create schemas for issuing and managing credentials of identities. It also provides a web-based [frontend (UI)](ui/README.md) to manage issuer schemas, credentials and connections.

This repository is for anyone to create their own [issuer node](https://0xpolygonid.github.io/tutorials/issuer-node/issuer-node-overview/) for Polygon ID.

---

## Installation

There are two options for installing and running the server alongside the UI:

1. [Docker Quick Start Guide](#docker-quick-start-guide)
2. [Standalone Mode Guide](#standalone-mode-guide)

### Docker Quick Start Guide

Running the app with Docker allows for minimal installation and a quick setup. This is recommended **for evaluation use-cases only**, such as local development builds.

**Full overview:**
1. Copy `.env-api.sample` as `.env-api` and `.env-issuer.sample` as `.env-issuer`. Please see the [configuration](#configuration) section for more details.
2. Run `make up`. This launches 3 containers with Postgres, Redis and Vault. Ignore the warnings about variables, since those are set up in the next step.
3. **If you are on an Apple Silicon chip (e.g. M1/M2), run `make run-arm`**. Otherwise, run `make run`. This starts up the issuer API, whose frontend can be accessed via the browser (default <http://localhost:3001>).
4. Follow the [steps](#adding-ethereum-private-key-to-the-vault) for adding an Ethereum private key to the Vault.
5. Follow the [steps](#creating-the-issuer-did) for creating an identity as your issuer DID.
6. _(Optional)_ To run the UI with its own API, first copy `.env-ui.sample` as `.env-ui`. Please see the [configuration](#configuration) section for more details.
7. _(Optional)_ Run `make run-ui` (or `make run-ui-arm` on Apple Silicon) to have the Web UI available on <http://localhost:8088> (in production mode). Its HTTP auth credentials are set in `.env-ui`. The UI API also has a frontend for API documentation (default <http://localhost:3002>).

#### Docker Guide Requirements

- Unix-based operating system (e.g. Debian, Arch, Mac OS)
- [Docker Engine](https://docs.docker.com/engine/) `1.27+`
- Makefile toolchain `GNU Make 3.81`

ℹ️ _**NOTE:** There is no compatibility with Windows environments at this time._

To help expedite a lot of the docker commands, a lot has been abstracted using make commands with the `Makefile`.
Included in each command are the docker commands to show what is being run.

#### Create Docker Configuration Files

Make a copy of the following environment variables files:

```bash
# FROM: ./

cp .env-api.sample .env-api;
cp .env-issuer.sample .env-issuer;
# (Optional - For Issuer UI)
cp .env-ui.sample .env-ui;
```

#### Node Issuer Configuration

The `.env-issuer` will be loaded into the [docker compose initiliaser](/infrastructure/local/docker-compose.yml)

You can use one of the following RPC providers:

- [Chainstack](https://chainstack.com/)
- [Ankr](https://ankr.com/)
- [QuickNode](https://quicknode.com/)
- [Alchemy](https://www.alchemy.com/)
- [Infura](https://www.infura.io/)

If you want to get up and running with a free public forwarding url, look at [Getting A Public URL](#getting-a-public-url).

**File:** `./.env-issuer`

```bash
# ...

ISSUER_SERVER_URL=https://unique-forwaring-or-public-url.ngrok-free.app
# Defaults for Basic Auth in Base64 ("user-issuer:password-issuer" = "dXNlci1pc3N1ZXI6cGFzc3dvcmQtaXNzdWVy")
# If you just want to get started, don't change these
ISSUER_API_AUTH_USER=user-issuer
ISSUER_API_AUTH_PASSWORD=password-issuer
# !!!MUST BE SET or other steps will not work
ISSUER_ETHEREUM_URL=<YOUR_RPC_PROVIDER_URI_ENDPOINT>
```

ℹ️ **NOTE:** In case you have loaded the vault multiple times and want a fresh start, run the following to remove remnant data:

```bash
# FROM: ./

rm ./infrastructure/local/.vault/data/init.out;
rm -rf ./infrastructure/local/.vault/file/*; # don't forget to type 'y'
rm ./infrastructure/local/.vault/plugins/vault-plugin-secrets-iden3;
rm -rf ./infrastructure/local/.vault/policies;

# Expected Output/Prompt:
#   sure you want to delete all 3 files in /Users/username/path/to/sh-id-platform/./infrastructure/local/.vault/file [yn]? y
```

#### Start Redis Postgres & Vault

This will start the necessary local services needed to store the wallet private key to the hashicord vault and allow storing data associated to the issuer.

```bash
# FROM: ./

make up;
# (Equivalent)
#   docker compose -p issuer -f ./infrastructure/local/docker-compose-infra.yml up -d redis postgres vault;

# Expected Output:
#   docker compose -p issuer -f /Users/username/path/to/sh-id-platform/infrastructure/local/docker-compose-infra.yml up -d redis postgres vault
#   [+] Running 4/4
#   ⠿ Network issuer-network       Created                                                                                   0.0s
#   ⠿ Container issuer-vault-1     Started                                                                                   0.5s
#   ⠿ Container issuer-redis-1     Started                                                                                   0.4s
#   ⠿ Container issuer-postgres-1  Started  
````

If you want to remove all the services (ignore the warnings):

```bash
# FROM: ./

make down; 
# (Equivalent)
#   docker compose -p issuer -f ./infrastructure/local/docker-compose-infra.yml down --remove-orphans -v;

# Expected Output:
#   docker compose -p issuer -f /Users/username/path/to/sh-id-platform/infrastructure/local/docker-compose-infra.yml down --remove-orphans
#   [+] Running 4/3
#   ⠿ Container issuer-postgres-1  Removed                                                                                   0.2s
#   ⠿ Container issuer-redis-1     Removed                                                                                   0.2s
#   ⠿ Container issuer-vault-1     Removed                                                                                   0.2s
#   ⠿ Network issuer-network       Removed                                                                                   0.0s
#   docker compose -p issuer -f /Users/username/path/to/sh-id-platform/infrastructure/local/docker-compose.yml down --remove-orphans
#   WARN[0000] The "DOCKER_FILE" variable is not set. Defaulting to a blank string. 
#   WARN[0000] The "DOCKER_FILE" variable is not set. Defaulting to a blank string. 
#   WARN[0000] The "DOCKER_FILE" variable is not set. Defaulting to a blank string. 
#   WARN[0000] The "DOCKER_FILE" variable is not set. Defaulting to a blank string.
```

#### Import Wallet Private Key To Vault

In order to secure our wallet private key so that the issuer can use it to issue claims/credentials, we'll store it in the hashicorp vault.

ℹ️ **NOTE:** Make sure the wallet you're providing has Testnet Matic to send transactions.

```bash
# FROM: ./

# Make sure to verify that the issuer-vault-1 is full initialised to avoid: "Error writing data to iden3/import/pbkey: Error making API request."
make private_key=YOUR_WALLET_PRIVATE_KEY add-private-key;
# (Equivalent)
#   docker exec issuer-vault-1 vault write iden3/import/pbkey key_type=ethereum private_key=YOUR_WALLET_PRIVATE_KEY;

# Expected Output:
#   docker exec issuer-vault-1 \
#           vault write iden3/import/pbkey key_type=ethereum private_key=YOUR_WALLET_PRIVATE_KEY
#   Success! Data written to: iden3/import/pbkey
```

#### **Add Vault To Configuration File**

This will get the vault token from the hashicorp vault docker instance and add it to our `./env-issuer` file.

```bash
# FROM: ./

make add-vault-token;
# (Equivalent)
#   TOKEN=$(docker logs issuer-vault-1 2>&1 | grep " .hvs" | awk  '{print $2}' | tail -1);
#	sed '/ISSUER_KEY_STORE_TOKEN/d' .env-issuer > .env-issuer.tmp;
#	echo ISSUER_KEY_STORE_TOKEN=$TOKEN >> .env-issuer.tmp;
#	mv .env-issuer.tmp .env-issuer;

# Expected Output:
#   sed '/ISSUER_KEY_STORE_TOKEN/d' .env-issuer > .env-issuer.tmp
#   mv .env-issuer.tmp .env-issuer
```

#### **Create Issuer DID**

This will create a new issuer DID by creating a new docker instance of the issuer, generating the DID for the issuer, storing it in the database, and deleting the instance.

It generates a new DID for the issuer node and saves it to `.env-api`

ℹ️ For _NON-Apple-M1/M2/Arm_ (ex: Intel/AMD):

```bash
# FROM: ./

# NON-Apple-M1/M2/Arm Command:
make generate-issuer-did;
# (Equivalent)
#   COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_FILE="Dockerfile" docker compose -p issuer -f ./infrastructure/local/docker-compose.yml up -d initializer
#	sleep 5
# 	$(eval DID = $(shell docker logs -f --tail 1 issuer-initializer-1 | grep "did"))
# 	@echo $(DID)
# 	sed '/ISSUER_API_UI_ISSUER_DID/d' .env-api > .env-api.tmp
# 	@echo ISSUER_API_UI_ISSUER_DID=$(DID) >> .env-api.tmp
# 	mv .env-api.tmp .env-api
# 	docker rm issuer-initializer-1
```

ℹ️ For _Apple-M1/M2/Arm_:

```bash
# FROM: ./

# Apple-M1/M2/Arm Command:
make generate-issuer-did-arm;
# (Equivalent)
#   COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_FILE="Dockerfile-arm" docker compose -p issuer -f /Users/username/path/to/sh-id-platform/infrastructure/local/docker-compose.yml up -d initializer;
#	sleep 5;
#   DID=$(docker logs -f --tail 1 issuer-initializer-1 | grep "did");
#   echo $DID;
#   sed '/ISSUER_API_UI_ISSUER_DID/d' .env-api > .env-api.tmp;
#   echo ISSUER_API_UI_ISSUER_DID=$DID >> .env-api.tmp;
#   mv .env-api.tmp .env-api;
#   docker rm issuer-initializer-1;

# Expected Output:
#   COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_FILE="Dockerfile-arm" docker compose -p issuer -f /Users/username/path/to/sh-id-platform/infrastructure/local/docker-compose.yml up -d initializer
#   WARN[0000] Found orphan containers ([issuer-vault-1 issuer-postgres-1 issuer-redis-1]) for this project. If you removed or renamed this service in your compose file, you can run this command with the --remove-orphans flag to clean it up. 
#   [+] Running 1/1
#   ⠿ Container issuer-initializer-1  Started                                                                                0.2s
#   sleep 5
#   did:polygonid:polygon:mumbai:uniqueAlphanumericKeyGenerated
#   sed '/ISSUER_API_UI_ISSUER_DID/d' .env-api > .env-api.tmp
#   mv .env-api.tmp .env-api
#   docker rm issuer-initializer-1
#   issuer-initializer-1
```

#### **Start Issuer Node API**

Now that you have all your files configured we can now start the issuer node api itself.

ℹ️ For _NON-Apple-M1/M2/Arm_ (ex: Intel/AMD):

```bash
# FROM: ./

make run-arm;
# (Equivalent)
#   COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_FILE="Dockerfile-arm" docker compose -p issuer -f /Users/username/path/to/sh-id-platform/infrastructure/local/docker-compose.yml up -d api;

# Expected Output:
#   COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_FILE="Dockerfile-arm" docker compose -p issuer -f /Users/username/path/to/sh-id-platform/local/docker-compose.yml up -d api;
#   WARN[0000] Found orphan containers ([issuer-vault-1 issuer-postgres-1 issuer-redis-1]) for this project. If you removed or renamed this service in your compose file, you can run this command with the --remove-orphans flag to clean it up. 
```

ℹ️ For _Apple-M1/M2/Arm_:

```bash
# FROM: ./

make run-arm;
# (Equivalent)
#   COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_FILE="Dockerfile-arm" docker compose -p issuer -f /Users/username/path/to/sh-id-platform/infrastructure/local/docker-compose.yml up -d api;

# Expected Output:
#   COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_FILE="Dockerfile-arm" docker compose -p issuer -f /Users/username/path/to/sh-id-platform/local/docker-compose.yml up -d api;
#   WARN[0000] Found orphan containers ([issuer-vault-1 issuer-postgres-1 issuer-redis-1]) for this project. If you removed or renamed this service in your compose file, you can run this command with the --remove-orphans flag to clean it up. 
```

Now open up [http://localhost:3001](http://localhost:3001) and see the following:

![Issuer Node API](/docs/3001.png)

---

### Standalone Mode Guide

Running the app in standalone mode means you will need to install the binaries for the server to run natively. This is essential for production deployments.

Make sure you have Postgres, Redis and Vault properly installed & configured. Do _not_ use `make up` since those will start the containers for non-production builds, see [Docker Quick Start Guide](#docker-quick-start-guide).

#### **Standalone Mode Guide Requirements**

- [Docker Engine](https://docs.docker.com/engine/) 1.27
- Makefile toolchain
- Unix-based operating system (e.g. Debian, Arch, Mac OS X)
- [Go](https://go.dev/) 1.19
- [Postgres](https://www.postgresql.org/)
- [Redis](https://redis.io/)
- [Hashicorp Vault](https://github.com/hashicorp/vault)

#### **Standalone Mode Setup**

1. Copy `.env-api.sample` as `.env-api` and `.env-issuer.sample` as `.env-issuer`. Please see the [configuration](#configuration) section for more details.
2. Run `make build`. This will generate a binary for each of the following commands:
    - `platform`
    - `migrate`
    - `pending_publisher`
    - `configurator`
3. Run `make db/migrate`. This checks the database structure and applies any changes to the database schema.
4. Follow the [steps](#adding-ethereum-private-key-to-the-vault) for adding an Ethereum private key to the Vault.
5. Run `./bin/platform` command to start the issuer.
6. Run `./bin/pending_publisher`. This checks that publishing transactions to the blockchain works.
7. Follow the [steps](#adding-ethereum-private-key-to-the-vault) for adding an Ethereum private key to the Vault.
8. Follow the [steps](#creating-the-issuer-did) for creating an identity as your issuer DID.
9. _(Optional)_ To set up the UI with its own API, first copy `.env-ui.sample` as `.env-ui`. Please see the [configuration](#configuration) section for more details.
10. _(Optional)_ Run `make run-ui` (or `make run-ui-arm` on Apple Silicon) to have the Web UI available on <http://localhost:8088> (in production mode). Its HTTP auth credentials are set in `.env-ui`. The UI API also has a frontend for API documentation (default <http://localhost:3002>).

> If you want to run the UI app in development mode, i.e. with HMR enabled, please follow the steps in the [Development (UI)](#development-ui) section.

---

## Issuing Credentials/Claims

ℹ️ **NOTE:** A lot of these steps can be done through the API UI at [http://localhost:3001](http://localhost:3001).

Once you've completed the [Installation](#installation) section, this will walk you through issuing credentials/claims.

### (Optional/Skip) Create Identity

This is to demonstrate how to createa an identity but is not needed as the identity is created in [Create Issuer DID](#create-issuer-did) step.

```bash
# NOTE: dXNlcjpwYXNzd29yZA== is a Basic HTTP Authorizatio as base64(user-issuer:password-issuer) from our config.toml file
# [HTTPBasicAuth]
# User="user"
# Password="password"
curl --location 'http://localhost:3001/v1/identities' \
--header 'Authorization: Basic dXNlci1pc3N1ZXI6cGFzc3dvcmQtaXNzdWVy' \
--header 'Content-Type: application/json' \
--data '{
    "didMetadata":{
        "method": "polygonid",
        "blockchain":"polygon",
        "network": "mumbai"
    }
}';

# Expected Output:
#   {"identifier":"did:polygonid:polygon:mumbai:2qPdb2hNczpXhkTDXfrNmmt9fGMzfDHewUnqGLahYE","state":{"claimsTreeRoot":"eb3d346d16f849b3cc2be69bfc58091dfaf6d90574be26bb40222aea67e08505","createdAt":"2023-03-22T22:49:02.782896Z","modifiedAt":"2023-03-22T22:49:02.782896Z","state":"b25cf54e7e648a263658416194c41ef6ae2dec101c50dfb2febc5e96eaa87110","status":"confirmed"}}
```

### (Optional/Skip) View Existing DIDs

This will output all dids that have been created with the issuer.

```bash
curl --location --request GET 'http://localhost:3001/v1/identities' \
--header 'Authorization: Basic dXNlci1pc3N1ZXI6cGFzc3dvcmQtaXNzdWVy' \
--header 'Content-Type: application/json' \
--data '{
    "did_metadata":{
        "method": "polygonid",
        "blockchain":"polygon",
        "network": "mumbai"
    }
}';

# Expected Output (Results may vary based on number of DIDs created):
#   ["did:polygonid:polygon:mumbai:2qMd3PtcVbzDNQZBiSDctaigbQtYW9KTqrLFoUm4Ur","did:polygonid:polygon:mumbai:2qMeNWv9xSSvWyBpn5tDojzQ8sga4VtrfuAkV65zQa","did:polygonid:polygon:mumbai:2qPdb2hNczpXhkTDXfrNmmt9fGMzfDHewUnqGLahYE","did:polygonid:polygon:mumbai:2qLR2qA22RemPeQDsQwdrrMU3SM9CNLnRBhmQtzo5v","did:polygonid:polygon:mumbai:2qHYtws8GQN3RniHLjPf5GuZUZtcD37o1MUgUmw287"]
```

### Creating Claim/Credentials

This will go through creating a `KYCAgeCredential` claim based off the following [KYC Age Credential Schema](https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json

Before creating the claim, the identifier of the service/person is needed. To retrive this, the `identifier` can be copied from the Polygon ID app to the clipboard.

!["ID Within Polygon ID App"](/docs/polygonid-app-id.png)

ℹ️ **NOTE:** The issuer node DID can be retrieved by looking at `.env-api` for `ISSUER_API_UI_ISSUER_DID`

```bash
curl --location 'http://localhost:3001/v1/did:polygonid:polygon:mumbai:2qPdb2hNczpXhkTDXfrNmmt9fGMzfDHewUnqGLahYE/claims' \
--header 'Authorization: Basic dXNlci1pc3N1ZXI6cGFzc3dvcmQtaXNzdWVy' \
--header 'Content-Type: application/json' \
--data '{
    "credentialSchema":"https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
    "type": "KYCAgeCredential",
    "credentialSubject": {
        "id": "did:polygonid:polygon:mumbai:2qEsg1AeTohAq6Euc3hBaDapfLVfQiWS7DUfvutYEq",
        "birthday": 19960424,
        "documentType": 2
    }
}';

# Expected Output:
#   {"id":"b1eab5be-dea3-11ed-8f7d-0242ac1e0005"}
```

### (Optional/Skip) Verifying Claim/Credentials Creation

Using the previous generated claim/credential id from [Creating Claim/Credentials](#creating-claimcredentials).

```bash
curl --location --request GET 'http://localhost:3001/v1/did:polygonid:polygon:mumbai:2qPdb2hNczpXhkTDXfrNmmt9fGMzfDHewUnqGLahYE/claims/b1eab5be-dea3-11ed-8f7d-0242ac1e0005' \
--header 'Authorization: Basic dXNlci1pc3N1ZXI6cGFzc3dvcmQtaXNzdWVy';

# Expected Output:
#   {"@context":["https://www.w3.org/2018/credentials/v1","https://schema.iden3.io/core/jsonld/iden3proofs.jsonld","https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld"],"credentialSchema":{"id":"https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json","type":"JsonSchemaValidator2018"},"credentialStatus":{"id":"https://unique-forwaring-or-public-url.ngrok-free.app/v1/did%3Apolygonid%3Apolygon%3Amumbai%3A2qPdb2hNczpXhkTDXfrNmmt9fGMzfDHewUnqGLahUQ/claims/revocation/status/2512063162","revocationNonce":2512063162,"type":"SparseMerkleTreeProof"},"credentialSubject":{"birthday":19960424,"documentType":2,"id":"did:polygonid:polygon:mumbai:2qEsg1AeTohAq6Euc3hBaDapfLVfQiWS7DUfvutYEq","type":"KYCAgeCredential"},"id":"http://localhost:3001/v1/did:polygonid:polygon:mumbai:2qPdb2hNczpXhkTDXfrNmmt9fGMzfDHewUnqGLahYE/claims/b1eab5be-dea3-11ed-8f7d-0242ac1e0005","issuanceDate":"2023-04-19T11:16:56.433871253Z","issuer":"did:polygonid:polygon:mumbai:2qPdb2hNczpXhkTDXfrNmmt9fGMzfDHewUnqGLahYE","proof":[{"type":"BJJSignature2021","issuerData":{"id":"did:polygonid:polygon:mumbai:2qPdb2hNczpXhkTDXfrNmmt9fGMzfDHewUnqGLahYE","state":{"claimsTreeRoot":"78b7651adb5d063553f7fdc11d279a3e307880aef6dec2b347abf0df53a11d27","value":"....
```

### Issuing Claim/Credential To Polygon ID App

In order to get the claim on the Polygon ID App, the claim QR Code payload is needed.

```bash
curl --location 'http://localhost:3001/v1/did:polygonid:polygon:mumbai:2qPdb2hNczpXhkTDXfrNmmt9fGMzfDHewUnqGLahYE/claims/b1eab5be-dea3-11ed-8f7d-0242ac1e0005/qrcode' \
--header 'Authorization: Basic dXNlci1pc3N1ZXI6cGFzc3dvcmQtaXNzdWVy';

# Expected Output:
#   {"body":{"credentials":[{"description":"https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld#KYCAgeCredential","id":"b1eab5be-dea3-11ed-8f7d-0242ac1e0005"}],"url":"http://localhost:3001/v1/agent"},"from":"did:polygonid:polygon:mumbai:2qPdb2hNczpXhkTDXfrNmmt9fGMzfDHewUnqGLahYE","id":"c01b66f1-d10b-4591-9dd6-8a94db1c112a","thid":"c01b66f1-d10b-4591-9dd6-8a94db1c112a","to":"did:polygonid:polygon:mumbai:2qEsg1AeTohAq6Euc3hBaDapfLVfQiWS7DUfvutYEq","typ":"application/iden3comm-plain-json","type":"https://iden3-communication.io/credentials/1.0/offer"}
```

Take this JSON data, copy, and paste into [https://qr.io](https://qr.io).

!["QR.io"](/docs/qrio.png)

With the Polygon ID app, open it up and scan the QR code.

!["Polygon ID App Adding Claim/Credential"](/docs/polygonid-app-claim.png)

### Verifying Claim

ℹ️ **NOTE:** The goal is to build your own type of claims and ways to verify claims, but this is an example of how things could work.

A quick way to validate this KYCAge Claim is to use [https://verifier-demo.polygonid.me/](https://verifier-demo.polygonid.me/).

!["Verifier Selecting KYCAgeCredential"](/docs/verifier-kycagecredential.png)

!["Verifier Verification Prompt"](/docs/verifier-verification.png)

!["Polygon ID App Generating Proof"](/docs/polygonid-app-proof.png)

!["Verifier Proof Verified"](/docs/verifier-success-verified.png)

---
## How To Run The Server

### Running the admin server backend (api) for evaluation purposes with docker-composer

1) Configure the project creating config.toml file.
2) Run `make up` to launch 3 containers with a postgres, redis and vault. This 3 containers are provided only for
   evaluation purposes.
3) Run `make run-ui-backend` to start a docker container running the issuer, (`make run-arm-ui-backend` for **Mac computers** with **Apple Silicon chip**)
4) Follow the [steps](#steps-to-write-the-private-key-in-the-vault) to write the private key in the vault
5) Browse to http://localhost:3002 (or the port configured in ServerAdminPort config entry)

---

## Configuration

For a full user guide, please refer to the [getting started docs](https://0xpolygonid.github.io/tutorials/issuer-node/getting-started-flow).

### Getting A Public URL

In order for the service to work, we'll need a public url.
An easy way to set this up is with using [ngrok](https://ngrok.com) as a forwarding service that maps to a local port.

```bash
# FROM: /path/to/ngrok binary

./ngrok http 3001;

# Expected Output:
# Add OAuth and webhook security to your ngrok (its free!): https://ngrok.com/free
# 
# Session Status                online
# Account                       YourAccountUsername (Plan: Free)
# Update                        update available (version 3.2.1, Ctrl-U to update)
# Version                       3.1.0
# Region                        Europe (eu)
# Latency                       -
# Web Interface                 http://127.0.0.1:4040
# Forwarding                    https://unique-forwading-address.eu.ngrok.io -> http://localhost:3001
# 
# Connections                   ttl     opn     rt1     rt5     p50     p90
                              # 0       0       0.00    0.00    0.00    0.00
```

### Advanced setup

Any variable defined in the config file can be overwritten using environment variables. The binding for this environment variables is defined in the function `bindEnv()` in the file `internal/config/config.go`

An _experimental_ helper command is provided via `make config` to allow an interactive generation of the config file, but this requires Go 1.19.

---

## Development (UI)

Completing either option of the [installation](#installation) process yields the UI as a minified Javascript app. Any changes to the UI source code would necessitate a Docker image removal and re-build to apply them. In most development scenarios this is undesirable, so the UI app can also be run in development mode like any [React](https://17.reactjs.org/) application to enable hot module replacement ([HMR](https://webpack.js.org/guides/hot-module-replacement/)).

1. Make sure that the UI API is set up and running properly (default <http://localhost:3002>).
2. Go to the `ui/` folder.
3. Copy the `.env.sample` file as `.env`
4. All variables are required to be set, with the exception of `VITE_ISSUER_LOGO`. The following are the corresponding variables present in the parent folder's `.env-api`, which need to be the same. Only `VITE_ISSUER_NAME` can differ for the UI to function in development mode.
    - `VITE_API_URL -> ISSUER_API_UI_SERVER_URL`
    - `VITE_API_USERNAME -> ISSUER_API_UI_AUTH_USER`
    - `VITE_API_PASSWORD -> ISSUER_API_UI_AUTH_PASSWORD`
    - `VITE_ISSUER_DID -> ISSUER_API_UI_ISSUER_DID`
    - `VITE_ISSUER_NAME -> ISSUER_API_UI_ISSUER_NAME`
    - `VITE_ISSUER_LOGO -> ISSUER_API_UI_ISSUER_LOGO`
5. Run `npm install`
6. Run `npm start`
7. The app will be running on <http://localhost:5173>.

---

## Testing

This will run you through the steps to test all aspects of the issuer node.

### Start Testing Environment

```bash
# FROM: ./sh-id-platform

make up-test;

# Expected Output:
# [+] Running 2/2
#  ⠿ Container sh-id-platform-test_postgres-1  Started                                                                                                                      0.3s
#  ⠿ Container sh-id-platform-test-vault       Running                                                                                                                      0.0s
```

### Run Tests

```bash
# FROM: ./sh-id-platform

# Run tests with race, use `make tests-race`
make tests;

# Expected Output:
# ...
# ?       github.com/polygonid/sh-id-platform/pkg/loaders [no test files]
# ?       github.com/polygonid/sh-id-platform/pkg/primitive       [no test files]
# ?       github.com/polygonid/sh-id-platform/pkg/protocol        [no test files]
# ?       github.com/polygonid/sh-id-platform/pkg/rand    [no test files]
# ?       github.com/polygonid/sh-id-platform/pkg/reverse_hash    [no test files]
# === RUN   TestMtSave
# --- PASS: TestMtSave (0.20s)
# PASS
# ok      github.com/polygonid/sh-id-platform/pkg/sync_ttl_map    0.549s
```

### Run Lint

```bash
# FROM: ./sh-id-platform

# Run tests with race, use `go test --race`
make lint;

# Expected Output:
# /path/to/sh-id-platform/bin/golangci-lint run
```

---

## Troubleshooting

In case any of the spun-up domains shows a 404 or 401 error when accessing their respective URLs, the root cause can usually be determined by inspecting the Docker container logs.

```bash
$ docker ps
CONTAINER ID   IMAGE                COMMAND                  CREATED          STATUS                    PORTS                                       NAMES
60e992ea9271   issuer-api-ui        "sh -c 'sleep 4s && …"   15 seconds ago   Up 4 seconds              0.0.0.0:3002->3002/tcp, :::3002->3002/tcp   issuer-api-ui-1
fae8873ac23b   issuer-ui            "/bin/sh /app/script…"   15 seconds ago   Up 4 seconds              0.0.0.0:8088->80/tcp, :::8088->80/tcp       issuer-ui-1
80d4511ed7c4   issuer-api           "sh -c 'sleep 4s && …"   21 minutes ago   Up 21 minutes             0.0.0.0:3001->3001/tcp, :::3001->3001/tcp   issuer-api-1
fa30b750848e   postgres:14-alpine   "docker-entrypoint.s…"   34 minutes ago   Up 34 minutes (healthy)   0.0.0.0:5432->5432/tcp, :::5432->5432/tcp   issuer-postgres-1
abd1d3c93255   redis:6-alpine       "docker-entrypoint.s…"   34 minutes ago   Up 34 minutes (healthy)   0.0.0.0:6379->6379/tcp, :::6379->6379/tcp   issuer-redis-1
0912c9920294   vault:latest         "docker-entrypoint.s…"   34 minutes ago   Up 34 minutes             0.0.0.0:8200->8200/tcp, :::8200->8200/tcp   issuer-vault-1
```

For example, for inspecting the issuer API node, run:

`docker logs issuer-api-1`

In most cases, a startup failure will be due to erroneous environment variables. In the case of the UI, any missing environment variable(s) will show as part of the error message.

---

## License

See [LICENSE](LICENSE.md).
