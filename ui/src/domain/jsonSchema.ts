export type Attribute =
  | StringAttribute
  | IntegerAttribute
  | NumberAttribute
  | BooleanAttribute
  | NullAttribute
  | ObjectAttribute
  | ArrayAttribute
  | MultiAttribute;

export type ArrayAttribute = {
  name: string;
  required: boolean;
  schema: ArraySchema;
  type: "array";
};

export type ArrayProps = {
  items?: Attribute;
};

export type ArraySchema = CommonProps & ArrayProps & { type: "array" };

export type BooleanAttribute = {
  name: string;
  required: boolean;
  schema: BooleanSchema;
  type: "boolean";
};

export interface BooleanProps {
  enum?: boolean[];
}

export type BooleanSchema = CommonProps & BooleanProps & { type: "boolean" };

export interface CommonProps {
  description?: string;
  title?: string;
}

export type IntegerSchema = CommonProps & { enum?: number[]; type: "integer" };

export type IntegerAttribute = {
  name: string;
  required: boolean;
  schema: IntegerSchema;
  type: "integer";
};

export type MultiAttribute = {
  name: string;
  required: boolean;
  schemas: MultiSchema;
  type: "multi";
};

export type MultiSchema = (CommonProps & (StringProps | BooleanProps | ObjectProps | ArrayProps))[];

export type NumberAttribute = {
  name: string;
  required: boolean;
  schema: NumberSchema;
  type: "number";
};

export type NullAttribute = {
  name: string;
  required: boolean;
  schema: NullSchema;
  type: "null";
};

export type NullSchema = CommonProps & { type: "null" };

export type NumberSchema = CommonProps & { enum?: number[]; type: "number" };

export type ObjectAttribute = {
  name: string;
  required: boolean;
  schema: ObjectSchema;
  type: "object";
};

export type ObjectProps = {
  properties?: Attribute[];
  required?: string[];
};

export type ObjectSchema = CommonProps & ObjectProps & { type: "object" };

export type JsonSchema = Attribute & SchemaProps;

export type SchemaProps = {
  $metadata: {
    uris: {
      jsonLdContext: string;
    };
  };
};

export type StringAttribute = {
  name: string;
  required: boolean;
  schema: StringSchema;
  type: "string";
};

export interface StringProps {
  enum?: string[];
  format?: string;
}

export type StringSchema = CommonProps & StringProps & { type: "string" };

export type JsonLdType = { id: string; name: string };
