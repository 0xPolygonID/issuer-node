export type IssuerIdentifier = string;

export enum IssuerType {
  BJJ = "BJJ",
  ETH = "ETH",
}

export enum AuthBJJCredentialStatus {
  "Iden3OnchainSparseMerkleTreeProof2023" = "Iden3OnchainSparseMerkleTreeProof2023",
  "Iden3ReverseSparseMerkleTreeProof" = "Iden3ReverseSparseMerkleTreeProof",
  "Iden3commRevocationStatusV1.0" = "Iden3commRevocationStatusV1.0",
}

export type Issuer = {
  authBJJCredentialStatus: AuthBJJCredentialStatus;
  blockchain: string;
  identifier: string;
  method: string;
  network: string;
};
