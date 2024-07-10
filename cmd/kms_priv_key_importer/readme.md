### Requirements
You have to have installed the following tools:
- [Go](https://golang.org/doc/install)
- [aws cli](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html) -- if you want to import your private key to AWS KMS
- [openssl](https://www.openssl.org/) -- if you want to import your private key to AWS KMS

this tools needs the following environment variables to be set:
```
# Could be either [localstorage | vault] (BJJ) and [localstorage | vault] | aws (ETH)
ISSUER_PUBLISH_KEY_PATH=pbkey
ISSUER_KMS_BJJ_PLUGIN=localstorage
ISSUER_KMS_ETH_PLUGIN=aws

# if the plugin is localstorage, you can specify the file path (default path is current directory)
# Important!!!: this path must be the same as the one used by the issuer node (defined in .env-issuer file)
ISSUER_KMS_PLUGIN_LOCAL_STORAGE_FILE_PATH=

# If the plugin is AWS for ETH keys you need to specify the key id and secret key
ISSUER_KMS_ETH_PLUGIN_AWS_ACCESS_KEY=XXX
ISSUER_KMS_ETH_PLUGIN_AWS_SECRET_KEY=YYY
ISSUER_KMS_ETH_PLUGIN_AWS_REGION=eu-west-1

# if the plugin is vault, you can specify the vault address and token
ISSUER_KEY_STORE_ADDRESS=http://localhost:8200
ISSUER_KEY_STORE_PLUGIN_IDEN3_MOUNT_PATH=iden3

# if the plugin is vault, you can specify the authentication method
ISSUER_VAULT_USERPASS_AUTH_ENABLED=false
ISSUER_VAULT_USERPASS_AUTH_PASSWORD=issuernodepwd
```

### How import your private key to AWS KMS

First you need to create a new key in AWS KMS, so export the variables defined in the requirements section and run the following command:

```
$ go run . --privateKey <private-key>
```
To get the key id you have to take a look at the output (or logs) of the previous command, it will be something like:

```logs
2024/07/10 10:38:14 INFO alias created: alias:=alias/pbkey
2024/07/10 10:38:14 INFO key created keyId=157a8b2a-e5e9-4414-b9c5-301ce828f6c5
```

then you can import your private key using the following command:

```shell
$ chmod +x kms_priv_key_importer
$ ./kms_priv_key_importer <private-key> <key-id> <aws-profile> <aws-region>
```

where:
* `private-key` is your private key in hex format
* `key-id` is the key id of the key created in AWS KMS (in this example `157a8b2a-e5e9-4414-b9c5-301ce828f6c5`)
* `aws-profile` is the profile name in your `~/.aws/credentials` file
* `aws-region` is the region where the key was created

if you get `Key material successfully imported!!!` message, then your private key was successfully imported to AWS KMS.

### Running Importer with Docker (AWS)
In the root project folder run:

```shell
docker build -t privadoid-kms-importer -f Dockerfile-kms-importer .
```

after the docker image is created run (make sure you have the .env-issuer with your env vars):

```shell
docker run -it -v ./.env-issuer:/.env-issuer privadoid-kms-importer sh
```

inside the container execute:

```
./kms_priv_key_importer privateKey <ETH-PRIV-KEY>
```

you will see something like this:

```shell
2024/07/10 15:27:54 INFO alias created: alias:=alias/pbkey
2024/07/10 15:27:54 INFO key created keyId=9bb5b78b-c288-44a7-b1d4-0543e0a6
```

then execute to set up AWS credentials:

```shell
aws configure set aws_access_key_id XXX --profile privado
aws configure set aws_secret_access_key YYY --profile privado
```
then import the material key

```shell
./aws_kms_material_key_importer.sh <ETH-PRIV-KEY> 9bb5b78b-c288-44a7-b1d4-0543e0a6 privado <AWS_REGION>
```
if you get `Key material successfully imported!!!` message, then your private key was successfully imported to AWS KMS.
