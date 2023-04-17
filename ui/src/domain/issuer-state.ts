export type TransactionStatus = "created" | "pending" | "published" | "failed";

export interface Transaction {
  id: number;
  publishDate: Date;
  state: string;
  status: TransactionStatus;
  txID: string;
}
