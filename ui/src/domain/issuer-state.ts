export type IssuerStatus = {
  pendingActions: boolean;
};

export type TransactionStatus = "created" | "pending" | "transacted" | "published" | "failed";

export type Transaction = {
  id: number;
  publishDate: Date;
  state: string;
  status: TransactionStatus;
  txID: string;
};
