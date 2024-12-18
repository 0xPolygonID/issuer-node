export type RouteID =
  | "connectionDetails"
  | "connections"
  | "credentialDetails"
  | "credentials"
  | "importSchema"
  | "issueCredential"
  | "issuerState"
  | "linkDetails"
  | "notFound"
  | "schemaDetails"
  | "schemas"
  | "identities"
  | "createIdentity"
  | "identityDetails"
  | "onboarding"
  | "displayMethods"
  | "displayMethodDetails"
  | "createDisplayMethod";

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
  createDisplayMethod: {
    layout: "sider",
    path: "/display-methods/create",
  },
  createIdentity: {
    layout: "sider",
    path: "/identities/create",
  },
  credentialDetails: {
    layout: "sider",
    path: "/credentials/issued/:credentialID",
  },
  credentials: {
    layout: "sider",
    path: "/credentials/:tabID",
  },
  displayMethodDetails: {
    layout: "sider",
    path: "/display-methods/:displayMethodID",
  },
  displayMethods: {
    layout: "sider",
    path: "/display-methods",
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
