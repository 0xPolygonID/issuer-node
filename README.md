[![Checks](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/checks.yml/badge.svg)](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/checks.yml)
[![golangci-lint](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/golangci-lint.yml)

# sh-id-platform

Polygon ID Self-Hosting Issuer Node Platform.
This repository is for anyone to create their own [issuer node](https://0xpolygonid.github.io/tutorials/issuer-node/issuer-node-overview/) for Polygon ID.

---

## Requirements

This project is developed with go 1.19. You probably would need a working golang environment to use it. 

You can also use the docker and docker compose files to run the project in case you do not want to compile by yoursef.

**Nice to have (Not all of them are strictly required but highly recommended):**
- Go `v1.19`
- Make `v3.81`
- Docker `v20.10.23` or greater
- Docker-compose.
- Unix style operating system (Linux, Mac, Windows WSL)

**Services needed for deployment:**
- Postgres
- Redis
- Hashicorp vault.

For your convenience, the testing environment can run these three services (Postgres, Redis and Vault)  in a docker 
for you. Please, consider that this is just for testing or evaluation purposes and _SHOULD NEVER_ be used in production without
securing it.

---

## Getting Started

This will walk you through on how to setup the services locally with Docker.

### Configuration File

Take note of the following things that will be changed.
We will get to these as we're configuring things.

**File:** `./config.toml.sample`

```toml
# Your issuer node exposed public URL for the Polygon ID to transact with
ServerUrl="<placeholder(public server url)>"

[KeyStore]
# The token generated from hashicorp/vault
Token="<placeholder(hvs.xxxxx from infrastructure/local/.vault/data/init.out file)>"

[Ethereum]
# Your RPC provider for Mumbai (Currently Mumbai is only supported)
URL="<placeholder(https://polygon-mumbai.g.alchemy.com/v2/xxxxxxxxx)>"
```

Make a copy of your `./config.toml.sample` file.

```bash
# FROM: ./sh-id-platform

cp ./config.toml.sample ./config.toml;
```

### Removing Docker Images

**NOTE:** This will be fixed in future versions, but make sure that that the following images are deleted

```bash
docker rmi sh-id-platform-platform:latest;
docker rmi vault:latest;
docker rmi postgres:14-alpine;
docker rmi redis:6-alpine;

# Expected Output:
# (Either image not found or the images are deleted)
```

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

Copy and paste that forwarding address into the `config.toml` file.

**File:** `./config.toml.sample`

```toml
ServerUrl="https://unique-forwading-address.eu.ngrok.io"
```

### RPC Provider

Using one of the following RPC providers get appropriate address to paste in the `config.toml` file.

**RPC Providers:**

- [Chainstack](https://chainstack.com)
- [Ankr](https://ankr.com)
- [QuickNode](https://quicknode.com)
- [Alchemy](https://www.alchemy.com)
- [Infura](https://www.infura.io)


**File:** `./config.toml.sample`

```toml
[Ethereum]
URL="https://your-mumbai-rpc.provider.address"
```

### Configuring Vault For Token

**NOTE:** This next step is optional, but in case you want to start from scratch and there is remaining vault data, use the following to remove any remnants of previous data, along with `make down`.

```bash
rm ./infrastructure/local/.vault/data/init.out;
rm -rf ./infrastructure/local/.vault/file/*;
rm ./infrastructure/local/.vault/plugins/vault-plugin-secrets-iden3;
rm -rf ./infrastructure/local/.vault/policies;
```

Start the your services to get the generated vault token.

```bash
# FROM: ./sh-id-platform

make up;

# Expected Output:
# WARN[0000] The "KEY_STORE_TOKEN" variable is not set. Defaulting to a blank string. 
# WARN[0000] The "KEY_STORE_TOKEN" variable is not set. Defaulting to a blank string. 
# WARN[0000] The "DOCKER_FILE" variable is not set. Defaulting to a blank string. 
# WARN[0000] The "KEY_STORE_TOKEN" variable is not set. Defaulting to a blank string. 
# WARN[0000] The "DOCKER_FILE" variable is not set. Defaulting to a blank string. 
# [+] Running 4/4
 # ⠿ Network sh-id-platform_default       Created                                                                                                                                      0.0s
#  ⠿ Container sh-id-platform-postgres-1  Started                                                                                                                                      0.3s
#  ⠿ Container sh-id-platform-redis-1     Started                                                                                                                                      0.3s
#  ⠿ Container sh-id-platform-test-vault  # Started                                                                                                                                      0.4s
```

Copy the `Initial Root Token` generated in your newly generated `init.out` file.

**File:** `./infrastructure/local/.vault/data/init.out`

```txt

...


Initial Root Token: hvs.uniqueK3y

...
```

**File:** `./config.toml.sample`

```toml
[KeyStore]
Token="hvs.uniqueK3y"
```

### Setting Your Private Key

This part requires you storing your wallet's private key within the hashicorp vault for future signed transactions.

**NOTE:** You will need some mumbai testnet tokens in your wallet for the service to work

```bash
docker exec -it sh-id-platform-test-vault sh;

# While in docker instance
$ vault write iden3/import/pbkey key_type=ethereum private_key=<YOUR-WALLET-PRIVATE-KEY>

# Expected Output:
# Success! Data written to: iden3/import/pbkey

$ exit;
```

### Running Services

Run the services for your respective OS.

```bash
# FROM: ./sh-id-platform

# Apple M1? use `make run-arm`
make run;

# Expected Output: (This might take a few minutes)
# ...
# [+] Running 4/4
#  ⠿ Container sh-id-platform-test-vault  Running                                                                                                                                      0.0s
#  ⠿ Container sh-id-platform-postgres-1  Running                                                                                                                                      0.0s
#  ⠿ Container sh-id-platform-redis-1     Running                                                                                                                                      0.0s
#  ⠿ Container sh-id-platform-platform-1  Started                                                                                                                                      0.3s
```

Double check that all services are running correctly.

```bash
docker ps -a;

# Expected Output:
# c659c0ee15c2   sh-id-platform-platform   "sh -c 'sleep 4s && …"   About a minute ago   Up 6 minutes                            sh-id-platform-platform-1
# e30d90c4df13   vault:latest              "docker-entrypoint.s…"   6 minutes ago        Up 6 minutes                    0.0.0.0:8200->8200/tcp   sh-id-platform-test-vault
# de8417dfa2ad   postgres:14-alpine        "docker-entrypoint.s…"   6 minutes ago        Up 6 minutes (healthy)          0.0.0.0:5432->5432/tcp   sh-id-platform-postgres-1
# 2310595c9d46   redis:6-alpine            "docker-entrypoint.s…"   6 minutes ago        Up 6 minutes (unhealthy)        0.0.0.0:6380->6379/tcp   sh-id-platform-redis-1
```

You can now go to `http://localhost:3001` and see the following working.


![https://localhost:3001](/docs/3001.png)

---

## Issuing Credentials/Claims

Once you've completed the [Getting Started](#getting-started) section, this will walk you through issuing credentials/claims.

### Create Identity

First step is to create an identity for our issuer.

```bash
# NOTE: dXNlcjpwYXNzd29yZA== is a Basic HTTP Authorizatio as base64(user:password) from our config.toml file
# [HTTPBasicAuth]
# User="user"
# Password="password"
curl --location 'http://localhost:3001/v1/identities' \
--header 'Authorization: Basic dXNlcjpwYXNzd29yZA==' \
--header 'Content-Type: application/json' \
--data '{
    "didMetadata":{
        "method": "polygonid",
        "blockchain":"polygon",
        "network": "mumbai"
    }
}';

# Expected Output
# {"identifier":"did:polygonid:polygon:mumbai:2qEeo6LxFqEcUEsqbc9hXUXq6PPSv3YHibinJyvY3N","state":{"claimsTreeRoot":"eb3d346d16f849b3cc2be69bfc58091dfaf6d90574be26bb40222aea67e08505","createdAt":"2023-03-22T22:49:02.782896Z","modifiedAt":"2023-03-22T22:49:02.782896Z","state":"b25cf54e7e648a263658416194c41ef6ae2dec101c50dfb2febc5e96eaa87110","status":"confirmed"}}
```

Let's verify the different ids that exist by retrieving all our identities.

```bash
curl --location --request GET 'http://localhost:3001/v1/identities' \
--header 'Authorization: Basic dXNlcjpwYXNzd29yZA==' \
--header 'Content-Type: application/json' \
--data '{
    "did_metadata":{
        "method": "polygonid",
        "blockchain":"polygon",
        "network": "mumbai"
    }
}';

# Expected Output:
# ["did:polygonid:polygon:mumbai:2qEeo6LxFqEcUEsqbc9hXUXq6PPSv3YHibinJyvY3N"]
```

### Creating Claim/Credentials

We're going to create a KYCAgeCredential claim base off the following (KYC Age Credential Schema)[https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json]

But before we can create a claim, we need to know the identity of the service/person we are generating this for. We can get the id from the Polygon ID and copy it to our clipboard.

!["ID Within Polygon ID App"](/docs/polygonid-app-id.png)

Using the following payload, replace the id with the id from your Polygon ID app, and make the request.

```bash
curl --location 'http://localhost:3001/v1/did:polygonid:polygon:mumbai:2qEeo6LxFqEcUEsqbc9hXUXq6PPSv3YHibinJyvY3N/claims' \
--header 'Authorization: Basic dXNlcjpwYXNzd29yZA==' \
--header 'Content-Type: application/json' \
--data '{
    "credentialSchema":"https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
    "type": "KYCAgeCredential",
    "credentialSubject": {
        "id": "did:polygonid:polygon:mumbai:2qEsg1BeTohAq7Euc4hBaDapfLVfQiWS6DSfvutWEq",
        "birthday": 19960424,
        "documentType": 2
    }
}';

# Expected Output:
# {"id":"6dd268dc-c906-11ed-9922-0242c0a82005"}
```

Let's double check that the claim has been successfully claimed.

```bash
curl --location --request GET 'http://localhost:3001/v1/did:polygonid:polygon:mumbai:2qEeo6LxFqEcUEsqbc9hXUXq6PPSv3YHibinJyvY3N/claims/6dd268dc-c906-11ed-9922-0242c0a82005' \
--header 'Authorization: Basic dXNlcjpwYXNzd29yZA==';

# Expected Output:
# {"@context":["https://www.w3.org/2018/credentials/v1","https://schema.iden3.io/core/jsonld/iden3proofs.jsonld","https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld"],"credentialSchema":{"id":"https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json","type":"JsonSchemaValidator2018"},"credentialStatus":{"id":"https://unique-forwading-address.eu.ngrok.io/v1/did%3Apolygonid%3Apolygon%3Amumbai%3A2qEeo6LxFqEcUEsqbc9hXUXq6PPSv3YHibinJyvY3N/claims/revocation/status/1609416217","revocationNonce":1609416217,"type":"SparseMerkleTreeProof"},"credentialSubject":{"birthday":19960424,"documentType":2,"id":"did:polygonid:polygon:mumbai:2qEsg1BeTohAq7Euc4hBaDapfLVfQiWS6DSfvutWEq","type":"KYCAgeCredential"},"id":"https://unique-forwading-address.eu.ngrok.io/v1/did:polygonid:polygon:mumbai:2qEeo6LxFqEcUEsqbc9hXUXq6PPSv3YHibinJyvY3N/claims/6dd268dc-c906-11ed-9922-0242c0a82005","issuanceDate":"2023-03-22T23:08:16.784637421Z","issuer":"did:polygonid:polygon:mumbai:2qEeo6LxFqEcUEsqbc9hXUXq6PPSv3YHibinJyvY3N","proof":[{"type":"BJJSignature2021","issuerData":{"id":"did:polygonid:polygon:mumbai:2qEeo6LxFqEcUEsqbc9hXUXq6PPSv3YHibinJyvY3N","state":{"claimsTreeRoot":"eb3d346d16f849b3cc2be69bfc58091dfaf6d90574be26bb40222aea67e08505","value":"....
```

### Issuing Claim/Credential To Polygon ID App

In order to get the claim on the Polygon ID App, we'll need to get the claim QR Code.

```bash
curl --location 'http://localhost:3001/v1/did:polygonid:polygon:mumbai:2qEeo6LxFqEcUEsqbc9hXUXq6PPSv3YHibinJyvY3N/claims/6dd268dc-c906-11ed-9922-0242c0a82005/qrcode' \
--header 'Authorization: Basic dXNlcjpwYXNzd29yZA==';

# Expected Output:
# {"body":{"credentials":[{"description":"https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld#KYCAgeCredential","id":"6dd268dc-c906-11ed-9922-0242c0a82005"}],"url":"https://unique-forwading-address.eu.ngrok.io/v1/agent"},"from":"did:polygonid:polygon:mumbai:2qEeo6LxFqEcUEsqbc9hXUXq6PPSv3YHibinJyvY3N","id":"d1c90cf3-f7b6-44b0-b4ce-f5bd9f9735b0","thid":"d1c90cf3-f7b6-44b0-b4ce-f5bd9f9735b0","to":"did:polygonid:polygon:mumbai:2qEsg1BeTohAq7Euc4hBaDapfLVfQiWS6DSfvutWEq","typ":"application/iden3comm-plain-json","type":"https://iden3-communication.io/credentials/1.0/offer"}
```

Take this JSON data, copy, and paste into (https://qr.io)[https://qr.io].

!["QR.io"](/docs/qrio.png)

With your phone that has the Polygon ID app installed on it, open it up and scan the QR code.

!["Polygon ID App Adding Claim/Credential"](/docs/polygonid-app-claim.png)

### Verifying Claim

**NOTE:** The goal is that you build your own type of claims and ways to verify claims, but this is an example of how you could see things working.

A quick way to validate this KYC Claim is to use [https://verifier-demo.polygonid.me/](https://verifier-demo.polygonid.me/).


!["Verifier Selecting KYCAgeCredential"](/docs/verifier-kycagecredential.png)

!["Verifier Verification Prompt"](/docs/verifier-verification.png)

!["Polygon ID App Generating Proof"](/docs/polygonid-app-proof.png)

!["Verifier Proof Verified"](/docs/verifier-success-verified.png)


---
## How To Run The Server

### Running the server for evaluation purposes with docker-composer

1) Configure the project creating config.toml file.
2) Run `make up` to launch 3 containers with a postgres, redis and vault. This 3 containers are provided only for
evaluation purposes. 
3) Run `make run` to start a docker container running the issuer, (`make run-arm` for **Mac computers** with **Apple Silicon chip**)
4) Follow the [steps](#steps-to-write-the-private-key-in-the-vault) to write the private key in the vault 
5) Browse to http://localhost:3001 (or the port configured in ServerPort config entry

### Running the admin server backend (api) for evaluation purposes with docker-composer

1) Configure the project creating config.toml file.
2) Run `make up` to launch 3 containers with a postgres, redis and vault. This 3 containers are provided only for
   evaluation purposes.
3) Run `make run-ui-backend` to start a docker container running the issuer, (`make run-arm-ui-backend` for **Mac computers** with **Apple Silicon chip**)
4) Follow the [steps](#steps-to-write-the-private-key-in-the-vault) to write the private key in the vault
5) Browse to http://localhost:3002 (or the port configured in ServerAdminPort config entry)

### Running the server in standalone mode

1) Configure the project creating a config.toml file copying the original config.toml.sample. The same variables can be
   injected as environment variables for your convenience. Please, see configuration section
2) Compile it with `make build`. You need a golang 1.19 environment to do it. make build will run a go install so
it will generate a binary for each of the commands:
    - platform
    - migrate
    - pending_publisher
    - configurator
3) Make sure you have postgres, redis and hashicorp vault properly configured. You could use `make up` command to start
a postgres, redis and vault redis container. Use this images only for evaluation purposes.
4) Make sure that your database is properly configured (step 1) and run `make db/migrate` command. This will check for the
current structure of the database and will apply needed change to create or update the database schema.
5) Write the vault token in the config.toml file, once the vault is initialized the token can be found in _infrastructure/local/.vault/data/init.out_ or in the logs of the vault container.
6) Run `./bin/platform` command to start the issuer. Browse to http://localhost:3001 (or the port configured in ServerPort config entry)
This will show you the api documentation.
7) Run `./bin/pending_publisher` in background. This process is not strictly necessary but highly recommended. 
It checks for possible errors publishing transactions on chain and try to resend it.
8) Follow the [steps](#steps-to-write-the-private-key-in-the-vault) to write the private key in the vault


---

## How to configure

The server can be configured with a config file and/or environment variables. There is a config.toml.sample file provided
as a reference. The system expects to have a config.toml file in the working directory. 

Any variable defined in the config file can be overwritten using environment variables. The binding 
for this environment variables is defined in the function bindEnv() in file internal/config/config.go

A helper command is provided under the command `make config` to help in the generation of the config file. 

### Steps to write the private key in the vault
1. docker exec -it sh-id-platform-test-vault sh
2. vault write iden3/import/pbkey key_type=ethereum private_key=<privkey>

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

# Run tests with race, use `go test --race`
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





1) Start the testing environment 
``make up-test``
2) Run tests
``make tests`` to run test or ``make test-race`` to run tests with go test --race
3) Run linter
``make lint``

---
