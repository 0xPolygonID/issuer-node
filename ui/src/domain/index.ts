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

export type JsonLdType = { id: string; name: string };

export type JsonLiteral = string | number | boolean | null;

export type Json = JsonLiteral | { [key: string]: Json } | Json[];
