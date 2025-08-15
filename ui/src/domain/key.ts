export enum KeyType {
  babyjubJub = "babyjubJub",
  ed25519 = "ed25519",
  secp256k1 = "secp256k1",
}

export type Key = {
  id: string;
  isAuthCredential: boolean;
  keyType: KeyType;
  name: string;
  publicKey: string;
};
