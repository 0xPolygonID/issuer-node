export type DataSchema = {
  credentialSchema: string;
  credentialSubject: {
    "Adhar-number": number;
    Age: number;
    id: string;
  };
  expiration: Date;
  mtProof: boolean;
  requestId: string;
  signatureProof: boolean;
  type: string;
};
