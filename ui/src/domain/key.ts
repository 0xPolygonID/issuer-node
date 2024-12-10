export enum KeyType {
  BJJ = "BJJ",
  ETH = "ETH",
}

export type Key = {
  id: string;
  isAuthCoreClaim: boolean;
  keyType: KeyType;
  publicKey: string;
};
