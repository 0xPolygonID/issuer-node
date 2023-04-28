export type CredentialsTabIDs = "issued" | "links";

export type ProofType = "MTP" | "SIG";

export interface Credential {
  createdAt: Date;
  credentialSubject: Record<string, unknown>;
  expired: boolean;
  expiresAt: Date | null;
  id: string;
  proofTypes: ProofType[];
  revNonce: number;
  revoked: boolean;
  schemaHash: string;
  schemaType: string;
  schemaUrl: string;
  userID: string;
}

export type LinkStatus = "active" | "inactive" | "exceeded";

export interface Link {
  active: boolean;
  createdAt: Date;
  credentialExpiration: Date | null;
  credentialSubject: Record<string, unknown>;
  expiration: Date | null;
  id: string;
  issuedClaims: number;
  maxIssuance: number | null;
  proofTypes: ProofType[];
  schemaHash: string;
  schemaType: string;
  schemaUrl: string;
  status: LinkStatus;
}
