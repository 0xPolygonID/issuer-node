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

export type CredentialsTabIDs = "issued" | "links";

export interface LinkAttribute {
  name: string;
  value: string;
}

export type LinkStatus = "active" | "inactive" | "exceeded";

export interface Link {
  active: boolean;
  attributes: LinkAttribute[];
  expiration?: Date;
  id: string;
  issuedClaims: number;
  maxIssuance?: number;
  schemaType: string;
  schemaUrl: string;
  status: LinkStatus;
}
