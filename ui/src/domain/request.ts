export type RequestsTabIDs = "Request";

export type Request = {
  Active: boolean;
  IssuerId: string;
  age: string;
  created_at: Date;
  credentialType: string;
  credential_type: string;
  id: string;
  modified_at: Date;
  proof_id: string;
  proof_type: string;
  requestDate: Date;
  requestId: string;
  requestType: string;
  request_status: string;
  revNonce: number;
  revoked: boolean;
  role_type: string;
  schemaID: string;
  source: string;
  status: string;
  userDID: string;
  verifier_status: string;
  wallet_status: string;
};

export type RequestVc = {
  Age: string;
  ProofId: string;
  ProofType: string;
  RequestType: string;
  RoleType: string;
  Source: string;
  schemaID: string;
  userDID: string;
};
export type VcResponse = {
  id: string;
  msg: string;
};

export type VerifyRequestResponse = {
  msg: string;
  status: boolean;
};
