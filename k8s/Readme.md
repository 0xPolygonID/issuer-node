

### Add Metamask private key

```bash
kubectl create secret generic private-key-secret --from-literal=private-key=XXXX
```

#### Important variables to setup

File: **issuer-node-api-configmap.yaml**

```yaml
# Set up public url for server api
ISSUER_SERVER_URL: "http://localhost:3001"

# RPC
ISSUER_ETHEREUM_URL: "https://polygon-mumbai.g.alchemy.com/v2/XXX"
ISSUER_ETHEREUM_CONTRACT_ADDRESS: "0x134B1BE3..."
ISSUER_ETHEREUM_RESOLVER_PREFIX: "polygon:mumbai" # or "polygon:main"
```

File: **issuer-node-api-ui-configmap.yaml**

```yaml
# Set up public url for server api ui
ISSUER_API_UI_SERVER_URL: "http://localhost:3002"
ISSUER_API_IDENTITY_METHOD: polygonid
ISSUER_API_IDENTITY_BLOCKCHAIN: polygon
ISSUER_API_IDENTITY_NETWORK: mumbai # or main
```

File: **issuer-node-ui-configmap.yaml**

```yaml
# Set up public url for server api ui
ISSUER_API_UI_SERVER_URL: "http://localhost:3002"
ISSUER_UI_BLOCK_EXPLORER_URL: "https://mumbai.polygonscan.com" # or "https://polygonscan.com"
```
