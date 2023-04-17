export interface Env {
  api: {
    password: string;
    url: string;
    username: string;
  };
  blockExplorer: string;
  issuer: {
    did: string;
    logo?: string;
    name: string;
  };
}
