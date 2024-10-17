# Overview

This is a helm chart for deploying Privado ID issuer node on Kubernetes.
To learn more about Privado ID issuer, see [this](https://0xpolygonid.github.io/tutorials/issuer/issuer-overview).

## Architecture

![Architecture diagram](resources/polygon-id-issuer-k8s-app-architecture.png)

# Installation

### Prerequisites

#### Set up command-line tools

Make sure you have these tools installed.

- [kubectl](https://kubernetes.io/docs/reference/kubectl/overview/)
- [helm](https://helm.sh/)

### Setup Ingress

The first step is update and modify the ingress.yaml file according to your requirements. The ingress.yaml file is located in the `ingress_sample` directory.
The issuer node works over HTTPS, so you need to provide a valid certificate.

### Setup volumes

How the volumes are set up depends on the cloud provider you are using, so you need to set up the volumes according to your cloud provider. Please, take a look at the [volumes](https://kubernetes.io/docs/concepts/storage/volumes/) documentation. You have to set up the volumes in the `vault-pv.yaml` and `postgres-pv.yaml` file.

### Configure the app with environment variables

To set up the app, you need to configure the following environment variables.
The ISSUER_RESOLVER_FILE is a base64 encoded string of the resolver file. You can take a look at the resolver file [here](../../resolvers_settings_sample.yaml)

```shell
export APP_INSTANCE_NAME=polygon-id-issuer              # Sample name for the application
export NAMESPACE=default                                # Namespace where you want to deploy the application
export UI_DOMAIN=ui.example.com                         # Domain for the UI.
export API_DOMAIN=api.example.com                       # Domain for the API.
export PRIVATE_KEY='YOUR PRIVATE KEY'                   # Private key of the wallet (Ethereum private key wallet).
export UIPASSWORD="my ui password"                      # Password for user: ui-user. This password is used when the user visit the ui.
export UI_INSECURE=true                                 # Set as true if the ui doesn't require basic auth. If this value true UIPASSWORD can be blank
export ISSUERNAME="My Issuer"                           # Issuer Name. This value is shown in the UI
export VAULT_PWD=password                               # Vault password to anable issuer node to connect with vault. Put the password you want to use.
export ISSUER_RESOLVER_FILE="cG9XYZ0K+"                 # Base64 encoded string of the resolver file. You can take a look at the resolver file [here](../../resolvers_settings_sample.yaml)
```

## Install the helm chart

```bash
helm install "$APP_INSTANCE_NAME" . \
--create-namespace --namespace "$NAMESPACE" \
--set namespace="$NAMESPACE" \
--set uidomain="$UI_DOMAIN" \
--set apidomain="$API_DOMAIN" \
--set privatekey="$PRIVATE_KEY" \
--set uiPassword="$UIPASSWORD" \
--set issuerName="$ISSUERNAME" \
--set privateKey="$PRIVATE_KEY" \
--set vaultpwd="$VAULT_PWD" \
--set issuerUiInsecure=$UI_INSECURE \
--set issuerResolverFile="$ISSUER_RESOLVER_FILE"
```
