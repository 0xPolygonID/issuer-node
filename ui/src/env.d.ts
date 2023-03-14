interface ImportMeta {
  readonly env: ImportMetaEnv;
}

interface ImportMetaEnv {
  readonly VITE_API_PASSWORD: string;
  readonly VITE_API_URL: string;
  readonly VITE_API_USERNAME: string;
  readonly VITE_ISSUER_DID: string;
  readonly VITE_ISSUER_LOGO: string;
  readonly VITE_ISSUER_NAME: string;
}
