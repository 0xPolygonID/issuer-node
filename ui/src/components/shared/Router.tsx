import { ComponentType } from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import { ConnectionDetails } from "src/components/connections/ConnectionDetails";
import { ConnectionsTable } from "src/components/connections/ConnectionsTable";

import { CredentialLinkQR } from "src/components/credentials/CredentialLinkQR";
import { Credentials } from "src/components/credentials/Credentials";
import { IssueCredential } from "src/components/credentials/IssueCredential";
import { IssuerState } from "src/components/issuer-state/IssuerState";
import { FullWidthLayout } from "src/components/layouts/FullWidthLayout";
import { SiderLayout } from "src/components/layouts/SiderLayout";
import { ImportSchema } from "src/components/schemas/ImportSchema";
import { SchemaDetails } from "src/components/schemas/SchemaDetails";
import { Schemas } from "src/components/schemas/Schemas";
import { NotFound } from "src/components/shared/NotFound";
import { Layout, ROUTES, RouteID } from "src/routes";
import { ROOT_PATH } from "src/utils/constants";

const COMPONENTS: Record<RouteID, ComponentType> = {
  connectionDetails: ConnectionDetails,
  connections: ConnectionsTable,
  credentialLinkQR: CredentialLinkQR,
  credentials: Credentials,
  importSchema: ImportSchema,
  issueCredential: IssueCredential,
  issuerState: IssuerState,
  notFound: NotFound,
  schemaDetails: SchemaDetails,
  schemas: Schemas,
};

export function Router() {
  const getLayoutRoutes = (currentLayout: Layout) =>
    Object.entries(ROUTES).reduce((acc: React.ReactElement[], [keyRoute, { layout, path }]) => {
      const componentsEntry = Object.entries(COMPONENTS).find(
        ([keyComponent]) => keyComponent === keyRoute
      );
      const Component = componentsEntry ? componentsEntry[1] : NotFound;

      return layout === currentLayout
        ? [...acc, <Route element={<Component />} key={path} path={path} />]
        : acc;
    }, []);

  return (
    <Routes>
      <Route element={<Navigate to={ROUTES.schemas.path} />} path={ROOT_PATH} />

      <Route element={<FullWidthLayout />}>{getLayoutRoutes("fullWidth")}</Route>

      <Route element={<FullWidthLayout background="bg-light" />}>
        {getLayoutRoutes("fullWidthGrey")}
      </Route>

      <Route element={<SiderLayout />}>{getLayoutRoutes("sider")}</Route>
    </Routes>
  );
}
