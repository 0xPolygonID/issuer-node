export enum KeyType {
  babyjubJub = "babyjubJub",
  secp256k1 = "secp256k1",
}

export type Key = {
  id: string;
  isAuthCredential: boolean;
  keyType: KeyType;
  name: string;
  publicKey: string;
};
