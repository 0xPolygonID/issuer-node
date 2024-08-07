import { AuthBJJCredentialStatus } from "./identifier";

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
  coreClaim: string;
  issuerData: Record<string, unknown>;
  signature: string;
  type: CredentialProofType;
};

export type CredentialSchema = {
  id: string;
  type: string;
};

export type CredentialStatus = {
  id: string;
  revocationNonce: number;
  type: AuthBJJCredentialStatus;
};

export type CredentialDetail = {
  "@context": string[];
  credentialSchema: CredentialSchema;
  credentialStatus: CredentialStatus;
  credentialSubject: {
    type: string;
  } & Record<string, unknown>;
  expirationDate: Date | null;
  id: string;
  issuanceDate: Date;
  issuer: string;
  proof: Proof[];
  refreshService: RefreshService | null;
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

export type RevocationStatus = {
  issuer?: {
    claimsTreeRoot?: string;
    revocationTreeRoot?: string;
    rootOfRoots?: string;
    state?: string;
  };
  mtp: {
    existence: boolean;
    node_aux?: {
      key?: string;
      value?: string;
    };
    siblings?: string[];
  };
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
