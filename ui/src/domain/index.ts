export interface Organization {
  did: string;
  displayName: string;
  legalName: string | null;
  logo: string;
  modifiedAt: Date;
}

export type TabsSchemasIDs = "archivedSchemas" | "mySchemas";

export interface StyleVariables {
  avatarBg: string;
  bgLight: string;
  borderColor: string;
  cyanBg: string;
  cyanColor: string;
  errorBg: string;
  errorColor: string;
  primaryBg: string;
  primaryColor: string;
  successColor: string;
  tagBg: string;
  tagBgSuccess: string;
  tagColor: string;
  textColor: string;
  textColorSecondary: string;
}

// Claims
export interface BooleanClaimFormAttribute {
  name: string;
  type: "boolean";
  value: boolean;
}
export interface DateClaimFormAttribute {
  name: string;
  type: "date";
  value: Date;
}
export interface NumberClaimFormAttribute {
  name: string;
  type: "number";
  value: number;
}
export interface SingleChoiceClaimFormAttribute {
  name: string;
  type: "singlechoice";
  value: number;
}

export type ClaimFormAttribute =
  | BooleanClaimFormAttribute
  | DateClaimFormAttribute
  | NumberClaimFormAttribute
  | SingleChoiceClaimFormAttribute;

export interface ClaimForm {
  attributes: ClaimFormAttribute[];
  claimLinkExpiration: Date | undefined;
  expirationDate: Date | undefined;
  limitedClaims: number | undefined;
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
  properties?: Schema[];
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
  items?: Schema;
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

type RootProps = {
  $id?: string;
  $schema?: string;
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

export type Schema = RootProps & Attribute;
