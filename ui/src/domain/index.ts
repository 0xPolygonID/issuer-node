export type { AppError } from "src/domain/error";

export type { Connection } from "src/domain/connection";

export type {
  Credential,
  CredentialsTabIDs,
  IssuedQRCode,
  Link,
  LinkStatus,
  ProofType,
  RefreshService,
  CredentialDetail,
  RevocationStatus,
} from "src/domain/credential";

export { CredentialProofType } from "src/domain/credential";

export type { Env } from "src/domain/env";

export type { IssuerStatus, Transaction, TransactionStatus } from "src/domain/issuer-state";

export type { Json, JsonObject, JsonLiteral } from "src/domain/json";

export type {
  ArrayAttribute,
  ArrayAttributeValue,
  Attribute,
  AttributeValue,
  BooleanAttribute,
  BooleanAttributeValue,
  BooleanProps,
  BooleanSchema,
  CommonProps,
  IntegerAttribute,
  IntegerAttributeValue,
  IntegerProps,
  IntegerSchema,
  JsonLdType,
  JsonSchema,
  MultiAttribute,
  MultiAttributeValue,
  Schema,
  MultiValue,
  NullAttribute,
  NullAttributeValue,
  NullSchema,
  NumberAttribute,
  NumberAttributeValue,
  NumberProps,
  NumberSchema,
  ObjectAttribute,
  ObjectAttributeValue,
  ObjectProps,
  JsonSchemaProps,
  StringAttribute,
  StringAttributeValue,
  StringProps,
  StringSchema,
} from "src/domain/jsonSchema";

export type { Schema as ApiSchema } from "src/domain/schema";

export type { IssuerIdentifier, Issuer, IssuerInfo, SupportedNetwork } from "src/domain/issuer";

export { IssuerType, AuthBJJCredentialStatus, Method } from "src/domain/issuer";
