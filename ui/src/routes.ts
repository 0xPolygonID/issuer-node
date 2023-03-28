export type RouteID =
  | "connectionDetails"
  | "connections"
  | "credentialLink"
  | "credentials"
  | "importSchema"
  | "issueCredential"
  | "notFound"
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
    path: "/connection/:connectionID",
  },
  connections: {
    layout: "sider",
    path: "/connections",
  },
  credentialLink: {
    layout: "fullWidthGrey",
    path: "/credential-link/:credentialID",
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
  notFound: {
    layout: "fullWidth",
    path: "/*",
  },
  schemas: {
    layout: "sider",
    path: "/schemas",
  },
};
