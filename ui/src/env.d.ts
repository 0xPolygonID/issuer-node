interface ImportMeta {
  readonly env: ImportMetaEnv;
}

interface ImportMetaEnv {
  readonly API_PASSWORD: string;
  readonly API_URL: string;
  readonly API_USERNAME: string;
  readonly ISSUER_DID: string;
  readonly ISSUER_LOGO: string;
  readonly ISSUER_NAME: string;
}
