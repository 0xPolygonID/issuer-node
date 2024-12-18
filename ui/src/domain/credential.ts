import { CredentialStatusType } from "src/domain/identity";

export type CredentialsTabIDs = "issued" | "links";

export enum ProofType {
  BJJSignature2021 = "BJJSignature2021",
  Iden3SparseMerkleTreeProof = "Iden3SparseMerkleTreeProof",
}

export type RefreshService = {
  id: string;
  type: "Iden3RefreshService2023";
};

export type DisplayMethod = {
  id: string;
  type: "Iden3BasicDisplayMethodv2";
};

export type CredentialStatus = {
  revocationNonce: number;
  type: CredentialStatusType;
};

export type Credential = {
  credentialStatus: CredentialStatus;
  credentialSubject: Record<string, unknown>;
  displayMethod: DisplayMethod | null;
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

export type IssuedMessage = {
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
