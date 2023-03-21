import { ComponentType } from "react";
import { Navigate, Route, Routes } from "react-router-dom";

import { Credentials } from "src/components/credentials/Credentials";
import { IssueCredential } from "src/components/credentials/IssueCredential";
import { ScanCredentialLink } from "src/components/credentials/ScanCredentialLink";
import { FullWidthLayout } from "src/components/layouts/FullWidthLayout";
import { SiderLayout } from "src/components/layouts/SiderLayout";
import { ImportSchema } from "src/components/schemas/ImportSchema";
import { SchemaDetails } from "src/components/schemas/SchemaDetails";
import { Schemas } from "src/components/schemas/Schemas";
import { NotFound } from "src/components/shared/NotFound";
import { Layout, ROUTES, RouteID } from "src/routes";
import { ROOT_PATH } from "src/utils/constants";

const COMPONENTS: Record<RouteID, ComponentType> = {
  credentialLink: ScanCredentialLink,
  credentials: Credentials,
  importSchema: ImportSchema,
  issueCredential: IssueCredential,
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
