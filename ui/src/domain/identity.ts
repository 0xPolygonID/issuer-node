export enum IdentityType {
  BJJ = "BJJ",
  ETH = "ETH",
}

export enum CredentialStatusType {
  "Iden3OnchainSparseMerkleTreeProof2023" = "Iden3OnchainSparseMerkleTreeProof2023",
  "Iden3ReverseSparseMerkleTreeProof" = "Iden3ReverseSparseMerkleTreeProof",
  "Iden3commRevocationStatusV1.0" = "Iden3commRevocationStatusV1.0",
  "SparseMerkleTreeProof" = "SparseMerkleTreeProof",
}

export enum Method {
  iden3 = "iden3",
  polygonid = "polygonid",
}

export type Network = {
  name: string;
  rhsMode: [CredentialStatusType, ...CredentialStatusType[]];
};

export type Blockchain = {
  name: string;
  networks: [Network, ...Network[]];
};

export type Identity = {
  blockchain: string;
  credentialStatusType: CredentialStatusType;
  displayName: string | null;
  identifier: string;
  method: Method;
  network: string;
};

export type IdentityDetails = {
  credentialStatusType: CredentialStatusType;
  displayName: string | null;
  identifier: string;
  keyType: IdentityType;
};
