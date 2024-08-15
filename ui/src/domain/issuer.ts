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

export enum Blockchain {
  polygon = "polygon",
  privado = "privado",
}

export enum PolygonNetwork {
  amoy = "amoy",
  mainnet = "mainnet",
}

export enum PrivadoNetwork {
  main = "main",
  test = "test",
}

export type Network = {
  [Blockchain.polygon]: PolygonNetwork;
  [Blockchain.privado]: PrivadoNetwork;
};

export enum Method {
  privado = "privado",
}

export type Issuer = {
  authBJJCredentialStatus: AuthBJJCredentialStatus;
  blockchain: Blockchain;
  displayName: string;
  identifier: string;
  method: Method;
  network: PolygonNetwork | PrivadoNetwork;
};

export type IssuerInfo = {
  authCoreClaimRevocationStatus: {
    type: AuthBJJCredentialStatus;
  };
  displayName: string;
  identifier: IssuerIdentifier;
  keyType: IssuerType;
};
