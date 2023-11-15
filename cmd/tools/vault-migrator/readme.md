## Vault Migrator Tool
With this tool you can export and import keys from Vault. this tool is useful when you want to migrate keys from one 
Vault to another, or you want to back up your keys.
Before you start, make sure you have the following:
- the vault address
- the vault token
- Golang or Docker installed

### Option 1. How to export and imports keys from/to Vault with Golang

Run the following command to export all the keys from Vault 
```bash
go run cmd/tools/vault-migrator/main.go -operation=export -output-file=keys.json -vault-token=your-vault-token -vault-addr=http://localhost:8200
```
the command above will export all keys from Vault to a file called **keys.json** in the **./keys** directory.

Alternatively, you can specify **a did** to export only the keys that match the did
```bash
go run cmd/tools/vault-migrator/main.go -operation=export -input-file=keys.json -vault-token=your-vault-token  -vault-addr=http://localhost:8200 -did=did:polygonid:polygon:mumbai:2qPHBiiu1wJN3rCMaaXwJpm9mNvuNqZZukzqS3V4Jg
```

### How to import keys to Vault
Running the following command will import all the keys from the file keys.json to Vault. 
Make sure you have the file keys.json in the current directory and the vault address and token are correct.
```bash
go run  cmd/tools/vault-migrator/main.go -operation=import -input-file=./keys/keys.json -vault-token=your-vault-token -vault-addr=http://localhost:8200
```

### Option 2. How to export and imports keys from/to Vault with Docker and Makefile
Another option is to use Docker and Makefile to export and import keys from Vault.
Make sure you have Docker installed and run the following ccommandommand to export all the keys from Vault.
The following commands work if you are running the issuer node with docker, but taking a look at the Makefile 
you can see how to run the commands with a remote Vault.

```bash
make vault_token=<your-vault-token> vault-export-keys
```

the command above will export all keys from Vault to a file called keys.json in the current directory.

for importing keys to Vault, run the following command:
```bash
make vault_token=xxx vault-import-keys
```
the command above will import all the keys from the file keys.json to Vault.