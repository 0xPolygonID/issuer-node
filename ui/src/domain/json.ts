export type JsonLiteral = string | number | boolean | null;

export type Json = JsonLiteral | { [key: string]: Json } | Json[];
