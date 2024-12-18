export type RouteID =
  | "connectionDetails"
  | "connections"
  | "credentialDetails"
  | "credentials"
  | "importSchema"
  | "issueCredential"
  | "createAuthCredential"
  | "issuerState"
  | "linkDetails"
  | "notFound"
  | "schemaDetails"
  | "schemas"
  | "identities"
  | "createIdentity"
  | "identityDetails"
  | "onboarding"
  | "keys"
  | "keyDetails"
  | "createKey";

export type Layout = "fullWidth" | "fullWidthGrey" | "sider";

type Routes = Record<
  RouteID,
  {
    layout: Layout;
    path: string;
  }
>;

export const ROUTES: Routes = {
  connectionDetails: {
    layout: "sider",
    path: "/connections/:connectionID",
  },
  connections: {
    layout: "sider",
    path: "/connections",
  },
  createAuthCredential: {
    layout: "sider",
    path: "/credentials/auth",
  },
  createIdentity: {
    layout: "sider",
    path: "/identities/create",
  },
  createKey: {
    layout: "sider",
    path: "/keys/create",
  },
  credentialDetails: {
    layout: "sider",
    path: "/credentials/issued/:credentialID",
  },
  credentials: {
    layout: "sider",
    path: "/credentials/:tabID",
  },
  identities: {
    layout: "sider",
    path: "/identities",
  },
  identityDetails: {
    layout: "sider",
    path: "/identities/:identityID",
  },
  importSchema: {
    layout: "sider",
    path: "/schemas/import-schema",
  },
  issueCredential: {
    layout: "sider",
    path: "/credentials/issue",
  },
  issuerState: {
    layout: "sider",
    path: "/issuer-state",
  },
  keyDetails: {
    layout: "sider",
    path: "/keys/:keyID",
  },
  keys: {
    layout: "sider",
    path: "/keys",
  },
  linkDetails: {
    layout: "sider",
    path: "/credentials/links/:linkID",
  },
  notFound: {
    layout: "fullWidth",
    path: "/*",
  },
  onboarding: {
    layout: "fullWidthGrey",
    path: "/onboarding",
  },
  schemaDetails: {
    layout: "sider",
    path: "/schemas/:schemaID",
  },
  schemas: {
    layout: "sider",
    path: "/schemas",
  },
};
