export type Env = {
  api: {
    password: string;
    url: string;
    username: string;
  };
  blockExplorerUrl: string;
  buildTag?: string;
  ipfsGatewayUrl: string;
  issuer: {
    logo: string;
    name: string;
  };
  schemaExplorerAndBuilderUrl?: string;
  warningMessage?: string;
};
