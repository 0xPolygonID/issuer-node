export type Identifier = string;

export enum IdentityType {
  BJJ = "BJJ",
  ETH = "ETH",
}

export enum CredentialStatusType {
  "Iden3OnchainSparseMerkleTreeProof2023" = "Iden3OnchainSparseMerkleTreeProof2023",
  "Iden3ReverseSparseMerkleTreeProof" = "Iden3ReverseSparseMerkleTreeProof",
  "Iden3commRevocationStatusV1.0" = "Iden3commRevocationStatusV1.0",
}

export enum Method {
  iden3 = "iden3",
  polygonid = "polygonid",
}

export type SupportedNetwork = {
  blockchain: string;
  networks: [string, ...string[]];
};

export type Identity = {
  blockchain: string;
  credentialStatusType: CredentialStatusType;
  displayName: string;
  identifier: string;
  method: Method;
  network: string;
};

export type IdentityDetails = {
  credentialStatusType: CredentialStatusType;
  displayName: string;
  identifier: Identifier;
  keyType: IdentityType;
};
