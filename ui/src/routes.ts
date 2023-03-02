export type RouteID =
  | "claimLink"
  | "createSchema"
  | "importSchema"
  | "issueClaim"
  | "notFound"
  | "schemas"
  | "signIn";

export type Layout = "fullWidth" | "fullWidthGrey" | "sider";

type Routes = Record<
  RouteID,
  {
    layout: Layout;
    path: string;
  }
>;

export const ROUTES: Routes = {
  claimLink: {
    layout: "fullWidthGrey",
    path: "/claim-link/:claimID",
  },
  createSchema: {
    layout: "sider",
    path: "/claiming/create-schema",
  },
  importSchema: {
    layout: "sider",
    path: "/schemas/import-schema",
  },
  issueClaim: {
    layout: "sider",
    path: "/schemas/issue-claim/:schemaID",
  },
  notFound: {
    layout: "fullWidth",
    path: "/*",
  },
  schemas: {
    layout: "sider",
    path: "/schemas/:tabID",
  },
  signIn: {
    layout: "fullWidth",
    path: "/sign-in",
  },
};
