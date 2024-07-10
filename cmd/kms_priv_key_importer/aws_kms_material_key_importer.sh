#!/usr/bin/env bash

set -e
set +x

raw_key=$1
key_id=$2
aws_profile=$3
aws_region=$4

ASN1_PRIV_KEY_HEADER="302e0201010420"
ASN1_SECP256K1_OID="a00706052b8104000a"
OUT_FILE="priv_key.pkcs8"

if [ -z "${raw_key}" ] || [ -z "${key_id}" ] || [ -z "${aws_profile}" ] || [ -z "${aws_region}" ]; then
  echo "Usage: $1 $0 <private_key> <key_id> <aws_profile> <aws_region>"
  exit 1
fi

openssl pkcs8 -topk8 -outform DER -nocrypt -inform DER -in <(echo "${ASN1_PRIV_KEY_HEADER} ${raw_key} ${ASN1_SECP256K1_OID}" | xxd -r -p) -out ${OUT_FILE} &>/dev/null
printf "private key successfully written to: %s\n" "${OUT_FILE}"

export KEY=`aws kms get-parameters-for-import --region ${aws_region} --profile ${aws_profile} \
--key-id ${key_id} \
--wrapping-algorithm RSAES_OAEP_SHA_256 \
--wrapping-key-spec RSA_2048 \
--query '{Key:PublicKey,Token:ImportToken}' \
--output text`

echo $KEY | awk '{print $1}' > PublicKey.b64
echo $KEY | awk '{print $2}' > ImportToken.b64

openssl enc -d -base64 -A -in PublicKey.b64 -out PublicKey.bin
openssl enc -d -base64 -A -in ImportToken.b64 -out ImportToken.bin

openssl pkeyutl \
-encrypt \
-in priv_key.pkcs8 \
-out EncryptedKeyMaterial.bin \
-inkey PublicKey.bin \
-keyform DER \
-pubin -encrypt -pkeyopt rsa_padding_mode:oaep -pkeyopt rsa_oaep_md:sha256


aws kms import-key-material --region ${aws_region} --profile ${aws_profile} \
--key-id ${key_id} \
--encrypted-key-material fileb://EncryptedKeyMaterial.bin \
--import-token fileb://ImportToken.bin \
--expiration-model KEY_MATERIAL_DOES_NOT_EXPIRE


printf "Key material successfully imported!!!\n"