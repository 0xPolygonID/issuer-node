export type RouteID =
  | "profile"
  | "connectionDetails"
  | "login"
  | "connections"
  | "credentialDetails"
  | "credentialIssuedQR"
  | "credentialLinkQR"
  | "credentials"
  | "importSchema"
  | "issueCredential"
  | "issuerState"
  | "linkDetails"
  | "notFound"
  | "schemaDetails"
  | "notification"
  | "request"
  | "createRequest";

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
  createRequest: {
    layout: "sider",
    path: "/request/create",
  },
  credentialDetails: {
    layout: "sider",
    path: "/credentials/issued/:credentialID",
  },
  credentialIssuedQR: {
    layout: "fullWidthGrey",
    path: "/credentials/scan-issued/:credentialID",
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
  login: {
    layout: "fullWidth",
    path: "/login",
  },
  notFound: {
    layout: "fullWidth",
    path: "/*",
  },
  notification: {
    layout: "sider",
    path: "/notification",
  },
  profile: {
    layout: "sider",
    path: "/profile",
  },
  request: {
    layout: "sider",
    path: "/request",
  },
  schemaDetails: {
    layout: "sider",
    path: "/schemas/:schemaID",
  },
};
