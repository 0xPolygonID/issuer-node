export interface CommonProps {
  description?: string;
  title?: string;
}

// Primitives

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

export interface IntegerProps {
  enum?: number[];
}

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

export interface NumberProps {
  enum?: number[];
}

export type NumberSchema = CommonProps & NumberProps & { type: "number" };

export type NumberAttribute = {
  name: string;
  required: boolean;
  schema: NumberSchema;
  type: "number";
};

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

// Non-primitives

type ArrayProps = {
  items?: Attribute;
};

type ArraySchema = CommonProps & ArrayProps & { type: "array" };

export type ArrayAttribute = {
  name: string;
  required: boolean;
  schema: ArraySchema;
  type: "array";
};

export type ObjectProps = {
  properties?: Attribute[];
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

export type MultiSchema =
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
  schemas: MultiSchema[];
  type: "multi";
};

// Schema

export type Attribute =
  | BooleanAttribute
  | IntegerAttribute
  | NullAttribute
  | NumberAttribute
  | StringAttribute
  | ArrayAttribute
  | ObjectAttribute
  | MultiAttribute;

export type SchemaProps = {
  $metadata: {
    uris: {
      jsonLdContext: string;
    };
  };
};

export type JsonSchema = Attribute & SchemaProps;

export type JsonLdType = { id: string; name: string };

// Values

type RequiredValue<T> = { required: true; value: T } | { required: false; value: T | undefined };

export type BooleanAttributeValue = {
  name: string;
  schema: BooleanSchema;
  type: "boolean";
} & RequiredValue<boolean>;

export type IntegerAttributeValue = {
  name: string;
  schema: IntegerSchema;
  type: "integer";
} & RequiredValue<number>;

export type NullAttributeValue = {
  name: string;
  schema: NullSchema;
  type: "null";
} & RequiredValue<null>;

export type NumberAttributeValue = {
  name: string;
  schema: NumberSchema;
  type: "number";
} & RequiredValue<number>;

export type StringAttributeValue = {
  name: string;
  schema: StringSchema;
  type: "string";
} & RequiredValue<string>;

export type ArrayAttributeValue = {
  name: string;
  schema: ArraySchema;
  type: "array";
} & (
  | { required: true; value: AttributeValue[] }
  | { required: false; value: AttributeValue[] | undefined }
);

export type ObjectAttributeValue = {
  name: string;
  schema: ObjectSchema;
  type: "object";
} & (
  | { required: true; value: AttributeValue[] }
  | { required: false; value: AttributeValue[] | undefined }
);

export type MultiValue =
  | BooleanAttributeValue
  | IntegerAttributeValue
  | NullAttributeValue
  | NumberAttributeValue
  | StringAttributeValue
  | ArrayAttributeValue
  | ObjectAttributeValue;

export type MultiAttributeValue = {
  name: string;
  schemas: MultiSchema[];
  type: "multi";
} & ({ required: true; value: MultiValue } | { required: false; value: MultiValue | undefined });

export type AttributeValue =
  | BooleanAttributeValue
  | IntegerAttributeValue
  | NullAttributeValue
  | NumberAttributeValue
  | StringAttributeValue
  | ArrayAttributeValue
  | ObjectAttributeValue
  | MultiAttributeValue;
