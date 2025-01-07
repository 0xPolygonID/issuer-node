export type PaymentConfig = {
  amount: string;
  paymentOptionID: number;
  recipient: string;
  signingKeyID: string;
};

export type PaymentOption = {
  config: Array<PaymentConfig>;
  createdAt: Date;
  description: string;
  id: string;
  issuerDID: string;
  modifiedAt: Date;
  name: string;
};

export type PaymentConfiguration = {
  ChainID: number;
  PaymentOption: {
    ContractAddress: string;
    Name: string;
    Type: string;
  };
  PaymentRails: string;
};

export type PaymentConfigurations = {
  [key: string]: PaymentConfiguration;
};
