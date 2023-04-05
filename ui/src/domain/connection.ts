import { Credential } from "src/domain/credential";

export interface Connection {
  createdAt: Date;
  credentials: Credential[];
  id: string;
  issuerID: string;
  userID: string;
}
