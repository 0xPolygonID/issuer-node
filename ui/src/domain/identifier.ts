export type Identifier = string | null;

export enum IssuerType {
  BJJ = "BJJ",
  ETH = "ETH",
}

export enum AuthBJJCredentialStatus {
  "Iden3OnchainSparseMerkleTreeProof2023" = "Iden3OnchainSparseMerkleTreeProof2023",
  "Iden3ReverseSparseMerkleTreeProof" = "Iden3ReverseSparseMerkleTreeProof",
  "Iden3commRevocationStatusV1.0" = "Iden3commRevocationStatusV1.0",
}

export type IssuerFormData =
  | {
      authBJJCredentialStatus: AuthBJJCredentialStatus;
      blockchain: string;
      method: string;
      network: string;
      type: IssuerType.BJJ;
    }
  | {
      authBJJCredentialStatus?: never;
      blockchain: string;
      method: string;
      network: string;
      type: Exclude<IssuerType, IssuerType.BJJ>;
    };

export type Issuer = Omit<IssuerFormData, "type" | "authBJJCredentialStatus"> & {
  authBJJCredentialStatus: AuthBJJCredentialStatus;
  identifier: string;
};

export type IssuerState = {
  address: string;
  identifier: string;
  state: {
    blockNumber?: number;
    blockTimestamp?: number;
    claimsTreeRoot: string;
    createdAt: Date;
    identifier?: string;
    modifiedAt: Date;
    previousState?: string;
    revocationTreeRoot?: string;
    rootOfRoots?: string;
    state: string;
    stateID?: number;
    status: string;
    txID?: string;
  };
};
