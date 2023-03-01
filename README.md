[![Checks](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/checks.yml/badge.svg)](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/checks.yml)
[![golangci-lint](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/golangci-lint.yml)

# sh-id-platform

## Prerequisites
This project is developed with go 1.19. You probably would need a working golang environment to use it. 

You can also use the docker and docker compose files to run the project in case you do not want to compile by yoursef.

Nice to have (Not all of them are strictly required but highly recommended):
- Go 1.19.
- Makefile.
- Docker.
- Docker-compose.
- Unix style operating system (Linux, Mac, Windows WSL)

Services needed:
- Postgres
- Redis
- Hashicorp vault.

For your convenience, the testing environment can run these three services (Postgres, Redis and Vault)  in a docker 
for you. Please, consider that this is just for testing or evaluation purposes and SHOULD never be used in production without
securing it.

## How to run the server

### Running the server for evaluation purposes with docker-composer
1) Configure the project creating config.toml file.
2) Run `make up` to launch 3 containers with a postgres, redis and vault. This 3 containers are provided only for
evaluation purposes. 
3) Run `make run` to start a docker container running the issuer, (`make run-arm` for **Mac computers** with **Apple Silicon chip**)
4) Follow the [steps](#steps-to-write-the-private-key-in-the-vault) to write the private key in the vault 
5) Browse to http://localhost:3001 (or the port configured in ServerPort config entry)


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


## How to configure
The server can be configured with a config file and/or environment variables. There is a config.toml.sample file provided
as a reference. The system expects to have a config.toml file in the working directory. 

Any variable defined in the config file can be overwritten using environment variables. The binding 
for this environment variables is defined in the function bindEnv() in file internal/config/config.go

A helper command is provided under the command `make config` to help in the generation of the config file. 

### Steps to write the private key in the vault
1. docker exec -it sh-id-platform-test-vault sh
2. vault write iden3/import/pbkey key_type=ethereum private_key=<privkey>

## How to test it
1) Start the testing environment 
``make up-test``
2) Run tests
``make tests`` to run test or ``make test-race`` to run tests with go test --race
3) Run linter
``make lint``
