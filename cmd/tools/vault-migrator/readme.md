## Vault Migrator Tool
With this tool you can export and import keys from Vault. this tool is useful when you want to migrate keys from one 
Vault to another, or you want to back up your keys.
Before you start, make sure you have the following:
- the vault address
- the vault token
- Golang or Docker installed

### Option 1. How to export keys from Vault with Golang

Run the following command to export all the keys from Vault 
```bash
go run cmd/tools/vault-migrator/main.go -operation=export -output-file=keys.json -vault-token=your-vault-token -vault-addr=http://localhost:8200
```
the command above will export all keys from Vault to a file called keys.json in the current directory.

Alternatively, you can specify **a did** to export only the keys that match the did
```bash
go run cmd/tools/vault-migrator/main.go -operation=export -input-file=keys.json -vault-token=your-vault-token  -vault-addr=http://localhost:8200 -did=did:polygonid:polygon:mumbai:2qPHBiiu1wJN3rCMaaXwJpm9mNvuNqZZukzqS3V4Jg
```

### How to import keys to Vault
Running the following command will import all the keys from the file keys.json to Vault. 
Make sure you have the file keys.json in the current directory and the vault address and token are correct.
```bash
go run  cmd/tools/vault-migrator/main.go -operation=import -input-file=keys.json -vault-token=your-vault-token -vault-addr=http://localhost:8200
```

### Option 2. How to export keys from Vault with Docker and Makefile
