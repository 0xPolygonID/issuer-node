export type Env = {
  api: {
    password: string;
    url: string;
    username: string;
  };
  blockExplorerUrl: string;
  issuer: {
    did: string;
    logo: string;
    name: string;
  };
  warningMessage?: string;
};
