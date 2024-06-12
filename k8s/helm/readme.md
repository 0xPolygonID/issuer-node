# Overview

This is a helm chart for deploying Privado iD issuer node on Kubernetes.
To learn more about Privado iD issuer, see [this](https://0xpolygonid.github.io/tutorials/issuer/issuer-overview).

## Architecture

![Architecture diagram](resources/polygon-id-issuer-k8s-app-architecture.png)

# Installation

### Prerequisites

#### Set up command-line tools

Make sure you have these tools installed.

- [kubectl](https://kubernetes.io/docs/reference/kubectl/overview/)
- [helm](https://helm.sh/)

### Configure the app with environment variables

To set up the app, you need to configure the following environment variables, but not all of them are required. You can choose to use ingress with domain names or not.
If you don't use ingress, you need to provide a public IP. If you use ingress, you need to provide a domain name.

```shell
export APP_INSTANCE_NAME=polygon-id-issuer              # Sample name for the application
export NAMESPACE=default                                # Namespace where you want to deploy the application
export APP_DOMAIN=app.example.com                       # Domain for the API UI. To use this INGRESS_ENABLED must be true
export UI_DOMAIN=ui.example.com                         # Domain for the UI. To use this INGRESS_ENABLED must be true
export API_DOMAIN=api.example.com                       # Domain for the API.To use this INGRESS_ENABLED must be true
export PRIVATE_KEY='YOUR PRIVATE KEY'                   # Private key of the wallet (Metamask private key wallet).
export PUBLIC_IP='YOUR PUBLIC IP'                       # Provide the PUBLIC IP if you have any otherwise leave this field.
export MAINNET=false                                    # Specify if the network is main, if this value is false issuer node will use amoy
export UIPASSWORD="my ui password"                      # Password for user: ui-user. This password is used when the user visit the ui.
export ISSUERNAME="My Issuer"                           # Issuer Name. This value is shown in the UI
export ISSUER_ETHERUM_URL="https://polygon-amoy.XXXX" # Blockchain RPC.
export INGRESS_ENABLED=true                             # If this value is false you must provide a STATIC_IP
export VAULT_PWD=password                               # Vault password.
export RHS_MODE=None                                    # Reverse Hash Service mode. Options: None, OnChain, OffChain
export RHS_URL="https://reverse-hash-service.com"       # Reverse Hash Service URL. Required if RHS_MODE is OffChain
```

## Install the helm chart with ingress and domain names

```bash
helm install "$APP_INSTANCE_NAME" . \
--create-namespace --namespace "$NAMESPACE" \
--set namespace="$NAMESPACE" \
--set appdomain="$APP_DOMAIN" \
--set uidomain="$UI_DOMAIN" \
--set apidomain="$API_DOMAIN" \
--set privatekey="$PRIVATE_KEY" \
--set mainnet="$MAINNET" \
--set uiPassword="$UIPASSWORD" \
--set issuerName="$ISSUERNAME" \
--set issuerEthereumUrl="$ISSUER_ETHERUM_URL" \
--set ingressEnabled="true" \
--set privateKey="$PRIVATE_KEY" \
--set vaultpwd="$VAULT_PWD" \
--set rhsMode="$RHS_MODE" \
--set rhsUrl="$RHS_URL"
```

In the code above, the PUBLIC_IP is not provided because is not needed when the ingress is enabled.
In this case `$APP_DOMAIN`, `$UI_DOMAIN` and `$API_DOMAIN` are used to create the ingress.
If the **rhsMode="OffChain"** is specified, **the rhsUrl must be provided**, but if the rhsMode="OnChain" or rshMode="None" is specified, the rhsUrl is not needed.
After a few minutes, the ingress should be ready, and you can access the app at the domain you specified.

## Install the helm chart with a public IP

```bash
helm install "$APP_INSTANCE_NAME" . \
--create-namespace --namespace "$NAMESPACE" \
--set privatekey="$PRIVATE_KEY" \
--set publicIP="$PUBLIC_IP" \
--set mainnet="$MAINNET" \
--set uiPassword="$UIPASSWORD" \
--set issuerName="$ISSUERNAME" \
--set issuerEthereumUrl="$ISSUER_ETHERUM_URL" \
--set ingressEnabled="true" \
--set privateKey="$PRIVATE_KEY" \
--set vaultpwd="$VAULT_PWD" \
--set rhsMode="$RHS_MODE" \
--set rhsUrl="$RHS_URL"
```

In the code above, the publicIP is provided because is needed when the ingress is not enabled. In this case `$APP_DOMAIN`, `$UI_DOMAIN` and `$API_DOMAIN` are not used.

After a few minutes, you can access the app by visiting:

- http://`$PUBLIC_IP`:30001 for the API
- http://`$PUBLIC_IP`:30002 for the API UI
- http://`$PUBLIC_IP`:30003 for the UI
