import { Credential } from "src/domain/credential";
import { List } from "src/utils/types";

export type Connection = {
  createdAt: Date;
  credentials: List<Credential>;
  id: string;
  issuerID: string;
  userID: string;
};
