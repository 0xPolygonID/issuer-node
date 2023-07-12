export type Json = JsonLiteral | { [key: string]: Json } | Json[];

export type JsonLiteral = string | number | boolean | null;

export type JsonObject = { [key: string]: JsonLiteral | JsonObject };
