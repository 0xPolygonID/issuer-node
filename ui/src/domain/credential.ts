export type CredentialsTabIDs = "issued" | "links";

export type ProofType = "MTP" | "SIG";

export type RefreshService = {
  id: string;
  type: "Iden3RefreshService2023";
};

export type Credential = {
  createdAt: Date;
  credentialSubject: Record<string, unknown>;
  expired: boolean;
  expiresAt: Date | null;
  id: string;
  proofTypes: ProofType[];
  refreshService: RefreshService | null;
  revNonce: number;
  revoked: boolean;
  schemaHash: string;
  schemaType: string;
  schemaUrl: string;
  userID: string;
};

export type IssuedQRCode = {
  qrCode: string;
  schemaType: string;
};

export type LinkStatus = "active" | "inactive" | "exceeded";

export type Link = {
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
};
