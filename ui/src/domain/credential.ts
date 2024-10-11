export type CredentialsTabIDs = "issued" | "links";

export enum CredentialProofType {
  BJJSignature2021 = "BJJSignature2021",
  Iden3SparseMerkleTreeProof = "Iden3SparseMerkleTreeProof",
}

export type ProofType = "MTP" | "SIG";

export type RefreshService = {
  id: string;
  type: "Iden3RefreshService2023";
};

export type Proof = {
  type: CredentialProofType;
};

export type Credential = {
  credentialSubject: Record<string, unknown>;
  expirationDate: Date | null;
  expired: boolean;
  id: string;
  issuanceDate: Date;
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
  schemaType: string;
  universalLink: string;
};

export type LinkStatus = "active" | "inactive" | "exceeded";

export type Link = {
  active: boolean;
  createdAt: Date;
  credentialExpiration: Date | null;
  credentialSubject: Record<string, unknown>;
  deepLink: string;
  expiration: Date | null;
  id: string;
  issuedClaims: number;
  maxIssuance: number | null;
  proofTypes: ProofType[];
  schemaHash: string;
  schemaType: string;
  schemaUrl: string;
  status: LinkStatus;
  universalLink: string;
};
