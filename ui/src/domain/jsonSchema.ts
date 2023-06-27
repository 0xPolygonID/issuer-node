export type CommonProps = {
  const?: unknown;
  default?: unknown;
  description?: string;
  examples?: unknown[];
  title?: string;
};

// Primitives

export type BooleanProps = {
  enum?: boolean[];
};

export type BooleanSchema = CommonProps & BooleanProps & { type: "boolean" };

export type BooleanAttribute = {
  name: string;
  required: boolean;
  schema: BooleanSchema;
  type: "boolean";
};

export type IntegerProps = {
  enum?: number[];
};

export type IntegerSchema = CommonProps & IntegerProps & { type: "integer" };

export type IntegerAttribute = {
  name: string;
  required: boolean;
  schema: IntegerSchema;
  type: "integer";
};

export type NullSchema = CommonProps & { type: "null" };

export type NullAttribute = {
  name: string;
  required: boolean;
  schema: NullSchema;
  type: "null";
};

export type NumberProps = {
  enum?: number[];
};

export type NumberSchema = CommonProps & NumberProps & { type: "number" };

export type NumberAttribute = {
  name: string;
  required: boolean;
  schema: NumberSchema;
  type: "number";
};

export type StringProps = {
  enum?: string[];
  format?: string;
};

export type StringSchema = CommonProps & StringProps & { type: "string" };

export type StringAttribute = {
  name: string;
  required: boolean;
  schema: StringSchema;
  type: "string";
};

// Non-primitives

type ArrayProps = {
  attribute?: Attribute;
};

type ArraySchema = CommonProps & ArrayProps & { type: "array" };

export type ArrayAttribute = {
  name: string;
  required: boolean;
  schema: ArraySchema;
  type: "array";
};

export type ObjectProps = {
  attributes?: Attribute[];
  properties?: Record<string, unknown>;
  required?: string[];
};

type ObjectSchema = CommonProps & ObjectProps & { type: "object" };

export type ObjectAttribute = {
  name: string;
  required: boolean;
  schema: ObjectSchema;
  type: "object";
};

// Multi-type

export type Schema =
  | BooleanSchema
  | IntegerSchema
  | NullSchema
  | NumberSchema
  | StringSchema
  | ArraySchema
  | ObjectSchema;

export type MultiAttribute = {
  name: string;
  required: boolean;
  schemas: Schema[];
  type: "multi";
};

// JSON Schema

export type Attribute =
  | BooleanAttribute
  | IntegerAttribute
  | NullAttribute
  | NumberAttribute
  | StringAttribute
  | ArrayAttribute
  | ObjectAttribute
  | MultiAttribute;

export type JsonSchemaProps = {
  $metadata: {
    uris: {
      jsonLdContext: string;
    };
  };
  $schema: string;
};

export type JsonSchema = ObjectAttribute & { jsonSchemaProps: JsonSchemaProps };

export type JsonLdType = { id: string; name: string };

// Values

type RequiredValue<T> = { required: true; value: T } | { required: false; value: T | undefined };

export type BooleanAttributeValue = BooleanAttribute & RequiredValue<boolean>;

export type IntegerAttributeValue = IntegerAttribute & RequiredValue<number>;

export type NullAttributeValue = NullAttribute & RequiredValue<null>;

export type NumberAttributeValue = NumberAttribute & RequiredValue<number>;

export type StringAttributeValue = StringAttribute & RequiredValue<string>;

export type ArrayAttributeValue = ArrayAttribute & RequiredValue<AttributeValue[]>;

export type ObjectAttributeValue = ObjectAttribute & RequiredValue<AttributeValue[]>;

export type MultiAttributeValue = MultiAttribute & RequiredValue<MultiValue>;

export type MultiValue =
  | BooleanAttributeValue
  | IntegerAttributeValue
  | NullAttributeValue
  | NumberAttributeValue
  | StringAttributeValue
  | ArrayAttributeValue
  | ObjectAttributeValue;

export type AttributeValue =
  | BooleanAttributeValue
  | IntegerAttributeValue
  | NullAttributeValue
  | NumberAttributeValue
  | StringAttributeValue
  | ArrayAttributeValue
  | ObjectAttributeValue
  | MultiAttributeValue;
