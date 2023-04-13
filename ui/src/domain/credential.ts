export interface BooleanCredentialFormAttribute {
  name: string;
  type: "boolean";
  value: boolean;
}

export interface Credential {
  attributes: {
    type: string;
  };
  createdAt: Date;
  expired?: boolean;
  expiresAt?: Date;
  id: string;
  revoked?: boolean;
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

export type CredentialsTabIDs = "issued" | "links";

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

export interface LinkAttributes {
  name: string;
  value: string;
}

export type LinkStatus = "active" | "inactive" | "exceeded";

export interface Link {
  active: boolean;
  attributes: LinkAttributes[];
  expiration?: Date;
  id: string;
  issuedClaims: number;
  maxIssuance?: number;
  schemaType: string;
  status: LinkStatus;
}
