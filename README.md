[![Checks](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/checks.yml/badge.svg)](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/checks.yml)
[![golangci-lint](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/0xPolygonID/sh-id-platform/actions/workflows/golangci-lint.yml)

# sh-id-platform

## Prerequisites
This project is developed with go 1.19. You probably would need a working golang environment to use it. 

You can also use the docker and docker compose files to run the project in case you do not want to compile by yoursef.

Nice to have:
- Go 1.19.
- Makefile.
- Docker.
- Docker-compose.

Services needed:
- Postgres
- Redis
- Hashicorp vault.

For your convenience, the testing environment can run this three services (Postgres, Redis and Vault)  in a docker 
for you. Please, consider that this is just for testing or evaluation purposes and SHOULD never be used in production without
securing it.

## How to run the server

### Running the server in standalone mode

1) Compile it with ``make build``. You need a golang 1.19 environment to do it
2) Configure the project creating a config.toml file copying the original config.toml.sample. You can also avoid using  
3) 



In order to run the server you should follow the next steps:

1. Start docker containers:
```
make up
```
This will start 3 docker co

2. Create the database and run migrations:
```
make db/migrate
```

3. Set up vault token:
Add or modify the key store configuration in the config.toml file:
```toml
[KeyStore]
    Address="http://localhost:8200"
    # In testing mode this value should be taken from ./infrastructure/local/vault/data/init.out
    Token="hvs.YxU2dLZljGpqLyPYu6VeYJde" 
    PluginIden3MountPath="iden3"
```


### Third party tools
If you want to execute the github actions locally visit https://github.com/nektos/act