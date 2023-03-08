#!/usr/bin/env bash
set -ueo pipefail

make down
sudo make clean-vault

make up
while ! docker exec sh-id-platform-test-vault test -e /root/.vault-token; do sleep 1; done
while ! docker exec sh-id-platform-test-vault vault secrets list | grep -qc iden3; do sleep 1; done

docker exec -i sh-id-platform-test-vault vault write iden3/import/pbkey key_type=ethereum private_key=d1aba427d045bcc57eb653cdfc532c67adda7e36304b982c93e72f3a77af4986

make run &
while ! curl --silent --no-progress-meter --fail http://localhost:3001 > /dev/null; do sleep 1; done

AUTHORIZATION="Basic $(echo -n 'user:password' | base64)"

echo 2>&1 "IDENTITY"
IDENTITY="$(
curl --no-progress-meter -H "authorization: $AUTHORIZATION" -H "accept: application/json" -H "content-type: application/json" \
 -X POST "http://localhost:3001/v1/identities" \
 -d '{"didMetadata":{"method":"polygonid","blockchain":"polygon","network":"mumbai"}}' \
| jq -r .identifier
)"
echo "Identity: $IDENTITY"

echo 2>&1 "CLAIM_ID"
CLAIM_ID="$(
curl --no-progress-meter -H "authorization: $AUTHORIZATION" -H "accept: application/json" -H "content-type: application/json" \
 -X POST "http://localhost:3001/v1/$IDENTITY/claims" \
 -d '{"credentialSchema":"https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json","type":"KYCAgeCredential","credentialSubject":{"id":"did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ","birthday":19960424,"documentType":2},"expiration":12345}' \
| jq -r .id
)"

echo 2>&1 "PUBLISH"
curl --no-progress-meter -H "authorization: $AUTHORIZATION" -H "accept: application/json" -H "content-type: application/json" \
 -X POST "http://localhost:3001/v1/$IDENTITY/state/publish" \
 -d ''

kill %1
wait
