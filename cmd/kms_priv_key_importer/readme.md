### Requirements
You have to have installed the following tools:
- [Go](https://golang.org/doc/install)
- [Docker](https://docs.docker.com/get-docker/)


this tools needs the following environment variables to be set up:
```
# Could be either [localstorage | vault | aws-sm] (BJJ) and [localstorage | vault | aws-sm |aws-kms] | aws (ETH)
ISSUER_PUBLISH_KEY_PATH=pbkey
ISSUER_KMS_ETH_PROVIDER=aws-sm or aws-kms

# if the plugin is localstorage, you can specify the file path (default path is current directory)
# Important!!!: this path must be the same as the one used by the issuer node (defined in .env-issuer file)
ISSUER_KMS_PROVIDER_LOCAL_STORAGE_FILE_PATH=./localstoragekeys

# If the plugin is AWS for ETH keys you need to specify the key id and secret key
ISSUER_KMS_ETH_PLUGIN_AWS_ACCESS_KEY=XXX
ISSUER_KMS_ETH_PLUGIN_AWS_SECRET_KEY=YYY
ISSUER_KMS_ETH_PLUGIN_AWS_REGION=eu-west-1
ISSUER_KMS_AWS_URL=<optional-aws-url>

# if the plugin is vault, you can specify the vault address and token
ISSUER_KEY_STORE_ADDRESS=http://localhost:8200
ISSUER_KEY_STORE_PLUGIN_IDEN3_MOUNT_PATH=iden3

# if the plugin is vault, you can specify the authentication method
ISSUER_VAULT_USERPASS_AUTH_ENABLED=false
ISSUER_VAULT_USERPASS_AUTH_PASSWORD=issuernodepwd
```

Instead of setting the environment variables you use the `.env-issuer` file, you can copy the `.env-issuer.example` 
file and rename it to `.env-issuer` and set the values of the environment variables there.
Then from the root project folder you can run the following command **(just for vault o localstorage)**:

```shell
$ go run cmd/kms_priv_key_importer/main.go --privateKey <privateETHKey>
````
or you can build the binary and run it:

```shell
$ go build -o kms_priv_key_importer cmd/kms_priv_key_importer/main.go
```

### How import your private key to AWS KMS
First you need to create a new key in AWS KMS, so export the variables defined in the requirements section:
```shell
export ISSUER_KMS_ETH_PROVIDER=aws-kms
export ISSUER_KMS_AWS_ACCESS_KEY=<aws-access-key>
export ISSUER_KMS_AWS_SECRET_KEY=<aws-secret-key>
export ISSUER_KMS_AWS_REGION=<aws-region>
export ISSUER_KMS_AWS_URL=<aws_endpoint_url> # optional
```
and run the following command:

```shell
$ ./kms_priv_key_importer --privateKey <privateETHKey>
```

if you get `key created keyId=<key-id>` message, then your private key was successfully imported to AWS KMS.


### How import your private key to AWS Secrets Manager
Export the variables defined in the requirements section:
```shell
export ISSUER_KMS_ETH_PROVIDER=aws-sm
export ISSUER_KMS_AWS_ACCESS_KEY=<aws-access-key>
export ISSUER_KMS_AWS_SECRET_KEY=<aws-secret-key>
export ISSUER_KMS_AWS_REGION=<aws-region>
export ISSUER_KMS_AWS_URL=<aws_endpoint_url> # optional
```
and run the following command:

```shell
$ ./kms_priv_key_importer --privateKey <privateETHKey>
```
that's it, your private key was successfully imported to AWS Secrets Manager.


### Docker Alternative Sample (localstorage) 📂
First you need to build the docker image:
```shell
$ docker build -t kms-priv-key-importer -f ./cmd/kms_priv_key_importer
```
Then you can run the docker container with the following command (at the root of the project):
```shell
$ mkdir localstoragekeys
 
$ docker run --rm -it \
-e ISSUER_PUBLISH_KEY_PATH=pbkey \
-e ISSUER_KMS_ETH_PROVIDER=localstorage \
-e ISSUER_KMS_PROVIDER_LOCAL_STORAGE_FILE_PATH=localstoragekeys \
-v $(pwd)/localstoragekeys:/localstoragekeys \
kms-priv-key-importer kms_priv_key_importer --privateKey <privateETHKey>
```

### Docker Alternative Sample (AWS-KMS) 🐳
First you need to build the docker image:
```shell
$ docker build -t kms-priv-key-importer -f ./cmd/kms_priv_key_importer
```

Then you can run the docker container with the following command:
```shell
$ docker run --rm -it \
-e ISSUER_PUBLISH_KEY_PATH=pbkey \
-e ISSUER_KMS_ETH_PROVIDER=aws-kms \
-e ISSUER_KMS_AWS_REGION=<aws-region> \                   # "local" if you are using localstack
-e ISSUER_KMS_AWS_ACCESS_KEY=<aws-access-key> \
-e ISSUER_KMS_AWS_SECRET_KEY=<aws-secret-key> \
#    -e ISSUER_KMS_AWS_URL=http://host.docker.internal:4566 \ # optional, use it if you are using localstack
kms-priv-key-importer kms_priv_key_importer --privateKey <privateETHKey> 
```

### Docker Alternative Sample (AWS-SM) 🐳
First you need to build the docker image:
```shell
$ docker build -t kms-priv-key-importer -f ./cmd/kms_priv_key_importer
```

Then you can run the docker container with the following command:
```shell
$ docker run --rm -it \
-e ISSUER_PUBLISH_KEY_PATH=pbkey \
-e ISSUER_KMS_ETH_PROVIDER=aws-sm \
-e ISSUER_KMS_AWS_REGION=<aws-region> \  # "local" if you are using localstack
-e ISSUER_KMS_AWS_ACCESS_KEY=<aws-access-key> \
-e ISSUER_KMS_AWS_SECRET_KEY=<aws-secret-key> \
#    -e ISSUER_KMS_AWS_URL=http://host.docker.internal:4566 \ # optional, use it if you are using localstack
kms-priv-key-importer kms_priv_key_importer --privateKey <privateETHKey>
```