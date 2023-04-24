export type RouteID =
  | "connectionDetails"
  | "connections"
  | "credentials"
  | "importSchema"
  | "issueCredential"
  | "issuerState"
  | "notFound"
  | "scanLink"
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
  issuerState: {
    layout: "sider",
    path: "/issuer-state",
  },
  notFound: {
    layout: "fullWidth",
    path: "/*",
  },
  scanLink: {
    layout: "fullWidthGrey",
    path: "/credentials/scan-link/:linkID",
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
