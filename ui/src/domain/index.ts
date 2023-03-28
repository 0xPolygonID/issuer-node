export interface Env {
  api: {
    password: string;
    url: string;
    username: string;
  };
  issuer: {
    did: string;
    logo?: string;
    name: string;
  };
}

export type TabsCredentialsIDs = "issued" | "links";

// Credentials
export interface BooleanCredentialFormAttribute {
  name: string;
  type: "boolean";
  value: boolean;
}
export interface DateCredentialFormAttribute {
  name: string;
  type: "date";
  value: Date;
}
export interface NumberCredentialFormAttribute {
  name: string;
  type: "number";
  value: number;
}
export interface SingleChoiceCredentialFormAttribute {
  name: string;
  type: "singlechoice";
  value: number;
}

export type CredentialFormAttribute =
  | BooleanCredentialFormAttribute
  | DateCredentialFormAttribute
  | NumberCredentialFormAttribute
  | SingleChoiceCredentialFormAttribute;

export interface CredentialForm {
  attributes: CredentialFormAttribute[];
  expiration: Date | undefined;
  linkAccessibleUntil: Date | undefined;
  linkMaximumIssuance: number | undefined;
}

// Schemas
export interface CommonProps {
  description?: string;
  title?: string;
}

export interface StringProps {
  enum?: string[];
  format?: string;
}

export type StringSchema = CommonProps & StringProps & { type: "string" };

export type StringAttribute = {
  name: string;
  required: boolean;
  schema: StringSchema;
  type: "string";
};

export type IntegerSchema = CommonProps & { enum?: number[]; type: "integer" };

export type IntegerAttribute = {
  name: string;
  required: boolean;
  schema: IntegerSchema;
  type: "integer";
};

export type NumberSchema = CommonProps & { enum?: number[]; type: "number" };

export type NumberAttribute = {
  name: string;
  required: boolean;
  schema: NumberSchema;
  type: "number";
};

export interface BooleanProps {
  enum?: boolean[];
}

export type BooleanSchema = CommonProps & BooleanProps & { type: "boolean" };

export type BooleanAttribute = {
  name: string;
  required: boolean;
  schema: BooleanSchema;
  type: "boolean";
};

export type NullSchema = CommonProps & { type: "null" };

export type NullAttribute = {
  name: string;
  required: boolean;
  schema: NullSchema;
  type: "null";
};

export type ObjectProps = {
  properties?: Attribute[];
  required?: string[];
};

export type ObjectSchema = CommonProps & ObjectProps & { type: "object" };

export type ObjectAttribute = {
  name: string;
  required: boolean;
  schema: ObjectSchema;
  type: "object";
};

export type ArrayProps = {
  items?: Attribute;
};

export type ArraySchema = CommonProps & ArrayProps & { type: "array" };

export type ArrayAttribute = {
  name: string;
  required: boolean;
  schema: ArraySchema;
  type: "array";
};

export type MultiSchema = (CommonProps & (StringProps | BooleanProps | ObjectProps | ArrayProps))[];

export type MultiAttribute = {
  name: string;
  required: boolean;
  schemas: MultiSchema;
  type: "multi";
};

export type Attribute =
  | StringAttribute
  | IntegerAttribute
  | NumberAttribute
  | BooleanAttribute
  | NullAttribute
  | ObjectAttribute
  | ArrayAttribute
  | MultiAttribute;

export type SchemaProps = {
  $metadata: {
    uris: {
      jsonLdContext: string;
    };
  };
};

export type Schema = Attribute & SchemaProps;

export type JsonLdType = { id: string; name: string };

export type JsonLiteral = string | number | boolean | null;
export type Json = JsonLiteral | { [key: string]: Json } | Json[];
