export enum IdentityType {
  BJJ = "BJJ",
  ETH = "ETH",
}

export enum Method {
  iden3 = "iden3",
  polygonid = "polygonid",
}

export type Network = {
  name: string;
  rhsMode: [string, ...string[]];
};

export type SupportedNetwork = {
  blockchain: string;
  networks: [Network, ...Network[]];
};

export type Identity = {
  blockchain: string;
  credentialStatusType: string;
  displayName: string | null;
  identifier: string;
  method: Method;
  network: string;
};

export type IdentityDetails = {
  credentialStatusType: string;
  displayName: string | null;
  identifier: string;
  keyType: IdentityType;
};
