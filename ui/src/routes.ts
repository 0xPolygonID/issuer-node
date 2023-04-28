export type RouteID =
  | "connectionDetails"
  | "connections"
  | "credentialDetails"
  | "credentialLinkQR"
  | "credentials"
  | "importSchema"
  | "issueCredential"
  | "issuedQR"
  | "issuerState"
  | "linkDetails"
  | "notFound"
  | "schemaDetails"
  | "schemas";

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
  credentialDetails: {
    layout: "sider",
    path: "/credentials/issued/:credentialID",
  },
  credentialLinkQR: {
    layout: "fullWidthGrey",
    path: "/credentials/scan-link/:linkID",
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
    path: "/credentials/issue/:schemaID?",
  },
  issuedQR: {
    layout: "fullWidthGrey",
    path: "/credentials/scan-issued/:credentialID",
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
  schemaDetails: {
    layout: "sider",
    path: "/schemas/:schemaID",
  },
  schemas: {
    layout: "sider",
    path: "/schemas",
  },
};
