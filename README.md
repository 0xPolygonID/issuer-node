# Polygon ID Issuer Node

[![Checks](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/checks.yml/badge.svg)](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/checks.yml)
[![golangci-lint](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/golangci-lint.yml)

This is a set of tools and APIs for issuers of zk-proof credentials, designed to be extensible. It allows an authenticated user to create schemas for issuing and managing credentials of identities. It also provides a web-based [frontend (UI)](ui/README.md) to manage issuer schemas, credentials and connections.

## Installation

There are two options for installing and running the server alongside the UI.

### Option 1 - Using Docker only

Running the app with Docker allows for minimal installation and a quick setup. This is recommended **for evaluation use-cases only**, such as local development builds.

#### Requirements for Docker-only

- [Docker Engine](https://docs.docker.com/engine/) 1.27+
- Makefile toolchain
- Unix-based operating system (e.g. Debian, Arch, Mac OS X)

_NB: There is no compatibility with Windows environments at this time._

#### Setup for Docker-only

1. Copy `.env-api.sample` as `.env-api` and `.env-issuer.sample` as `.env-issuer`. Please see the [configuration](#configuration) section for more details.
2. Run `make up`. This launches 3 containers with Postgres, Redis and Vault. Ignore the warnings about variables, since those are set up in the next step.
3. **If you are on an Apple Silicon chip (e.g. M1/M2), run `make run-arm`**. Otherwise, run `make run`. This starts up the issuer API, whose admin frontend can be accessed via the browser (default <http://localhost:3001>).
4. Follow the [steps](#adding-ethereum-private-key-to-the-vault) for adding an Ethereum private key to the Vault.
5. Follow the [steps](#creating-the-issuer-did) for creating an identity as your issuer DID.
6. _(Optional)_ To run the UI with its own API, first copy `.env-ui.sample` as `.env-ui`. Please see the [configuration](#configuration) section for more details.
7. _(Optional)_ Run `make run-ui` (or `make run-ui-arm` on Apple Silicon) to have the Web UI available on <http://localhost:8088>. Its HTTP auth credentials are set in `.env-ui`. The UI API also has a frontend for API documentation (default <http://localhost:3002>).

### Option 2 - Standalone mode

Running the app in standalone mode means you will need to install the binaries for the server to run natively. This is essential for production deployments.

#### Requirements for standalone mode

- [Docker Engine](https://docs.docker.com/engine/) 1.27
- Makefile toolchain
- Unix-based operating system (e.g. Debian, Arch, Mac OS X)
- [Go](https://go.dev/) 1.19
- [Postgres](https://www.postgresql.org/)
- [Redis](https://redis.io/)
- [Hashicorp Vault](https://github.com/hashicorp/vault)

_NB: There is no compatibility with Windows environments at this time._

#### Setup for standalone mode

Make sure you have Postgres, Redis and Vault properly installed & configured. Do _not_ use `make up` since those will start the containers for non-production builds, see [option 1](#option-1---using-docker-only).

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
10. _(Optional)_ Run `make run-ui` (or `make run-ui-arm` on Apple Silicon) to have the Web UI available on <http://localhost:8088>. Its HTTP auth credentials are set in `.env-ui`. The UI API also has a frontend for API documentation (default <http://localhost:3002>).

## Configuration

For a full user guide, please refer to the [getting started docs](https://0xpolygonid.github.io/tutorials/issuer-node/getting-started-flow).

### Turnkey Docker-only setup

If you are setting up [locally](#setup-for-docker-only) with Docker, you will need to set up the following variables in their respective `.env` files.

In `.env-api`:

- `ISSUER_API_UI_AUTH_USER`
- `ISSUER_API_UI_AUTH_PASSWORD`
- `ISSUER_API_UI_ISSUER_DID` - obtained from [here](#creating-the-issuer-did).
- `ISSUER_ETHEREUM_URL` - this is the Ethereum address of the issuer's DApp.
- `ISSUER_API_UI_ISSUER_LOGO` - optional (placeholder used if left blank). A valid URL to a minimum 40x40 pixel PNG, JPEG or SVG of the issuer's logo.

In `.env-issuer`:

- `ISSUER_API_AUTH_USER`
- `ISSUER_API_AUTH_PASSWORD`
- `ISSUER_KEY_STORE_TOKEN` - obtained from step 3 [here](#adding-ethereum-private-key-to-the-vault).

If you are running the UI, in `.env-ui`:

- `ISSUER_UI_AUTH_USERNAME`
- `ISSUER_UI_AUTH_PASSWORD`

### Adding Ethereum private key to the Vault

This is required for signing on-chain transactions. In a basic use-case this can be retrieved from an Ethereum wallet that can connect to Polygon Mumbai Testnet.

Follow these steps:

1. Copy your Ethereum private key, pasting it into `<private_key>` in step 2.
2. Run `make private_key=<private_key> add-private-key`
3. Run `docker logs issuer-vault-1`
4. Look for the end of the output starting with `===== ENABLED IDEN3 =====`, and copy `<X>` in `token:<X>`.
5. Paste that token into the `ISSUER_KEY_STORE_TOKEN` variable in `.env-issuer`.

### Creating the issuer DID

This determines the owner of the credentials that are issued. You can either reuse an existing DID, or you can generate a new identity by following these steps:

1. Navigate to the issuer API frontend (default <http://localhost:3001>).
2. Under "AUTHENTICATION", enter the credentials provided in `.env-api` (`ISSUER_API_AUTH_USER` & `ISSUER_API_UI_AUTH_PASSWORD`), then click "SET".
3. Under "IDENTITY" > "Create Identity", click "TRY".
4. In the resulting JSON response, copy the "identity" value to `ISSUER_API_UI_ISSUER_DID` in `.env-api`.

### Advanced setup

Any variable defined in the config file can be overwritten using environment variables. The binding for this environment variables is defined in the function `bindEnv()` in the file `internal/config/config.go`

An _experimental_ helper command is provided via `make config` to allow an interactive generation of the config file, but this requires Go 1.19.

## Development

[TODO]

- Developing the issuer API
- Developing the UI API
- Developing the UI (HMR enabled with env config)

## Testing

Start the testing environment with `make up-test`

- Run tests with `make tests` to run test or `make test-race` to run tests with the Go parameter `test --race`
- Run the linter with `make lint`

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

In most cases, a startup failure will be due to erroneous env variables.

## License

See [LICENSE](LICENSE.md).
