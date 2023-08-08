
### Add Metamask private key and Vault password
* You must provide a private key for the Metamask account that will be used to sign transactions.
* You must provide a password for the Vault that will be used to store the private key and the issuer identifier.
```bash
export METAMASK_PRIVATE_KEY=XXXX
export VAULT_PASSWORD=XXXX
kubectl create secret generic private-key-secret \ 
	--from-literal=vault-pwd=$VAULT_PASSWORD \ 
	--from-literal=private-key=$METAMASK_PRIVATE_KEY
```

### Basic Configuration
Inside the folder `./mumbai` or `./mainnet` you will find the following files:
* ingress-patch.yaml
* issuer-node-api-configmap-path.yaml
* issuer-node-api-ui-configmap-path.yaml
* issuer-node-ui-deployment.yaml

In those file you must change your custom values according to your needs.

### Install Issuer Node - Mumbai network

```bash
kubectl apply -k ./mumbai
```
### Install Issuer Node - Main network

```bash
kubectl apply -k ./mainet
```
