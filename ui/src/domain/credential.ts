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
  expired: boolean;
  expiresAt?: Date;
  id: string;
  revNonce: number;
  revoked: boolean;
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
