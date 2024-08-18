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

export enum Method {
  iden3 = "iden3",
  polygonid = "polygonid",
}

export type SupportedNetwork = {
  blockchain: string;
  networks: [string, ...string[]];
};

export type Issuer = {
  authBJJCredentialStatus: AuthBJJCredentialStatus;
  blockchain: string;
  displayName: string;
  identifier: string;
  method: Method;
  network: string;
};

export type IssuerInfo = {
  authCoreClaimRevocationStatus: {
    type: AuthBJJCredentialStatus;
  };
  displayName: string;
  identifier: IssuerIdentifier;
  keyType: IssuerType;
};
