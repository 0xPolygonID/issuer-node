export interface IssuerStatus {
  arePendingActions: boolean;
}

export type TransactionStatus = "created" | "pending" | "transacted" | "published" | "failed";

export interface Transaction {
  id: number;
  publishDate: Date;
  state: string;
  status: TransactionStatus;
  txID: string;
}
