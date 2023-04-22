export interface Credential {
  createdAt: Date;
  credentialSubject: {
    type: string;
  };
  expired: boolean;
  expiresAt: Date | null;
  id: string;
  revNonce: number;
  revoked: boolean;
}

export type CredentialsTabIDs = "issued" | "links";

export type LinkStatus = "active" | "inactive" | "exceeded";

export type ProofType = "SIG" | "MTP";

export interface Link {
  active: boolean;
  expiration: Date | null;
  id: string;
  issuedClaims: number;
  maxIssuance: number | null;
  proofTypes: ProofType[];
  schemaType: string;
  schemaUrl: string;
  status: LinkStatus;
}
