#### Certificate Installation (Cloudflare + Let's Encrypt)

1. Install cert-manager
```shell
helm install cert-manager jetstack/cert-manager --namespace cert-manager --version v1.14.24 --set installCRDs=true
```

2. Create a secret with the Cloudflare API token:
Edit the `secret.yaml` file with your Cloudflare API token. Then run:
```shell
kubectl apply -f secret.yaml
```

3. Install the Cluster Issuer
```shell
kubectl apply -f issuer-prod.yaml
```

4. Install the certificate
Edit the `namespace` in the `certificate-prod.yaml` file and then run:
```shell
kubectl apply -f certificate-prod.yaml
```
