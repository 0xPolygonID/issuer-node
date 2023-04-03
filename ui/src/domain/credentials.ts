export interface BooleanCredentialFormAttribute {
  name: string;
  type: "boolean";
  value: boolean;
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

export type CredentialsTabIDs = "issued" | "links";
