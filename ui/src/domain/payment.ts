export type PaymentConfig = {
  amount: string;
  paymentOptionID: number;
  recipient: string;
  signingKeyID: string;
};

export type PaymentOption = {
  createdAt: Date;
  description: string;
  id: string;
  issuerDID: string;
  modifiedAt: Date;
  name: string;
  paymentOptions: Array<PaymentConfig>;
};

export type PaymentConfiguration = {
  ChainID: number;
  PaymentOption: {
    ContractAddress: string;
    Decimals: number;
    Name: string;
    Type: string;
  };
  PaymentRails: string;
};

export type PaymentConfigurations = {
  [key: string]: PaymentConfiguration;
};
