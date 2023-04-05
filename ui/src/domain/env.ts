export interface Env {
  api: {
    password: string;
    url: string;
    username: string;
  };
  issuer: {
    did: string;
    logo?: string;
    name: string;
  };
}
