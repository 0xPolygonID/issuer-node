export interface Env {
  api: {
    password: string;
    url: string;
    username: string;
  };
  blockExplorerUrl: string;
  buildTag?: string;
  issuer: {
    did: string;
    logo: string;
    name: string;
  };
  warningMessage?: string;
}
