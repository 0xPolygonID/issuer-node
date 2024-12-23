export type Env = {
  api: {
    password: string;
    url: string;
    username: string;
  };
  baseUrl?: string;
  buildTag?: string;
  displayMethodBuilderUrl: string;
  ipfsGatewayUrl: string;
  issuer: {
    logo: string;
    name: string;
  };
  schemaExplorerAndBuilderUrl?: string;
  warningMessage?: string;
};
