export type RouteID =
  | "connectionDetails"
  | "connections"
  | "credentialDetails"
  | "credentialIssuedQR"
  | "credentials"
  | "importSchema"
  | "issueCredential"
  | "issuerState"
  | "linkDetails"
  | "notFound"
  | "schemaDetails"
  | "schemas"
  | "issuers"
  | "createIssuer"
  | "issuerDetails"
  | "onboarding";

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
  createIssuer: {
    layout: "sider",
    path: "/issuers/create",
  },
  credentialDetails: {
    layout: "sider",
    path: "/credentials/issued/:credentialID",
  },
  credentialIssuedQR: {
    layout: "fullWidthGrey",
    path: "/credentials/scan-issued/:credentialID",
  },
  credentials: {
    layout: "sider",
    path: "/credentials/:tabID",
  },
  importSchema: {
    layout: "sider",
    path: "/schemas/import-schema",
  },
  issueCredential: {
    layout: "sider",
    path: "/credentials/issue",
  },
  issuerDetails: {
    layout: "sider",
    path: "/issuers/details/:issuerID",
  },
  issuers: {
    layout: "sider",
    path: "/issuers",
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
