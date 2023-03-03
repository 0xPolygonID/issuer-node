# Self-hosted Polygon ID Issuer Node

[![Checks](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/checks.yml/badge.svg)](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/checks.yml)
[![golangci-lint](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/golangci-lint.yml)

This is a set of tools and an API for issuers of zk-proof credentials to build on top of. It allows a single authenticated user to create schemas for issuing and managing credentials of identities. It also provides a web-based [frontend (UI)](ui/README.md) to manage issuer schemas, credentials and connections.

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

1. Copy `config.toml.sample` as `config.toml`. Please see the [configuration](#configuration) section for more details.
    - For a turnkey setup, you will only need to add a valid JSON RPC URL for Polygon Mumbai under `[Ethereum] URL`.
2. Run `make up`. This launches 3 containers with Postgres, Redis and Vault. Ignore the warnings about variables, since those are set up in the next step.
3. **If you are on an Apple Silicon chip (e.g. M1/M2), run `make run-arm`**. Otherwise, run `make run`. This starts Docker containers for the issuer application.
4. Follow the [steps](#adding-ethereum-private-key-to-the-vault) for adding an Ethereum private key to the Vault.
5. Open <http://localhost:3001> in a browser (or whatever was set in the `[Server] URL` config entry). This shows an admin interface for documentation and credentials issuer setup.
6. _(Optional)_ Run `make run-ui-backend` and `make-run-ui` to have the Platform UI available on <http://localhost:5173> (or whatever was set in the `[UI] URL` config entry).
7. MISSING STEP ON SETTING UP UI

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

1. Copy `config.toml.sample` as `config.toml`. Please see the [configuration](#configuration) section for more details.
2. Run `make build`. This will generate a binary for each of the following commands:
    - `platform`
    - `migrate`
    - `pending_publisher`
    - `configurator`
3. Run `make db/migrate`. This checks the database structure and applies any changes to the database schema.
4. Check the file `infrastructure/local/.vault/data/init.out` for the Vault token and copy it under `[KeyStore] Token` in `config.toml`.
5. Run `./bin/platform` command to start the issuer.
6. Run `./bin/pending_publisher`. This checks that publishing transactions to the blockchain works.
7. Follow the [steps](#adding-ethereum-private-key-to-the-vault) for adding an Ethereum private key to the Vault.
8. Open <http://localhost:3001> in a browser (or whatever was set in the `[Server] URL` config entry). This shows an admin interface for documentation and credentials issuer setup.
9. MISSING STEP ON SETTING UP UI IN PRODUCTION

## Configuration

For a full guide, please refer to the [getting started docs](https://0xpolygonid.github.io/tutorials/issuer-node/getting-started-flow).

The application requires a config file `config.toml` file in the root working directory. There is a `config.toml.sample` file provided as a reference.

Any variable defined in the config file can be overwritten using environment variables. The binding for this environment variables is defined in the function `bindEnv()` in the file `internal/config/config.go`

An _experimental_ helper command is provided via `make config` to allow an interactive generation of the config file, but this requires Go 1.19.

### Adding Ethereum private key to the Vault

This allows signing on-chain transactions. In a basic use-case this can be retrieved from an Ethereum wallet that can connect to Polygon Mumbai Testnet.

Run the following commands, then exit the CLI:

1. `docker exec -it sh-id-platform-test-vault sh`
2. `vault write iden3/import/pbkey key_type=ethereum private_key=<private_key>`

## Testing

Start the testing environment with `make up-test`

- Run tests with `make tests` to run test or `make test-race` to run tests with the Go parameter `test --race`
- Run the linter with `make lint`

## License

See [LICENSE](LICENSE.md).
