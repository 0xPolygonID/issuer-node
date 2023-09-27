export type RequestsTabIDs = "Request";

export type Request = {
  credentialType: string;
  requestDate: Date;
  requestId: string;
  requestType: string;
  status: string;
  userDID: string;
};
