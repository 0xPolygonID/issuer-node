import { ComponentType } from "react";
import { Navigate, Route, Routes } from "react-router-dom";

import { ConnectionDetails } from "src/components/connections/ConnectionDetails";
import { ConnectionsTable } from "src/components/connections/ConnectionsTable";
import { CredentialDetails } from "src/components/credentials/CredentialDetails";
import { Credentials } from "src/components/credentials/Credentials";
import { IssueCredential } from "src/components/credentials/IssueCredential";
import { LinkDetails } from "src/components/credentials/LinkDetails";
import { CreateIdentity } from "src/components/identities/CreateIdentity";
import { Identities } from "src/components/identities/Identities";
import { Identity } from "src/components/identities/Identity";
import { Onboarding } from "src/components/identities/Onboarding";
import { IssuerState } from "src/components/issuer-state/IssuerState";
import { FullWidthLayout } from "src/components/layouts/FullWidthLayout";
import { SiderLayout } from "src/components/layouts/SiderLayout";
import { ImportSchema } from "src/components/schemas/ImportSchema";
import { SchemaDetails } from "src/components/schemas/SchemaDetails";
import { Schemas } from "src/components/schemas/Schemas";
import { NotFound } from "src/components/shared/NotFound";
import { useIdentityContext } from "src/contexts/Identity";
import { Layout, ROUTES, RouteID } from "src/routes";
import { ROOT_PATH } from "src/utils/constants";

const COMPONENTS: Record<RouteID, ComponentType> = {
  connectionDetails: ConnectionDetails,
  connections: ConnectionsTable,
  createIdentity: CreateIdentity,
  credentialDetails: CredentialDetails,
  credentials: Credentials,
  identities: Identities,
  identityDetails: Identity,
  importSchema: ImportSchema,
  issueCredential: IssueCredential,
  issuerState: IssuerState,
  linkDetails: LinkDetails,
  notFound: NotFound,
  onboarding: Onboarding,
  schemaDetails: SchemaDetails,
  schemas: Schemas,
};

export function Router() {
  const { identifier } = useIdentityContext();

  const filteredRoutes = identifier
    ? Object.entries(ROUTES).filter(([, { path }]) => path !== ROUTES.onboarding.path)
    : Object.entries(ROUTES).filter(([, { path }]) => path === ROUTES.onboarding.path);

  const getLayoutRoutes = (currentLayout: Layout) =>
    filteredRoutes.reduce((acc: React.ReactElement[], [keyRoute, { layout, path }]) => {
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
      <Route
        element={<Navigate to={identifier ? ROUTES.schemas.path : ROUTES.onboarding.path} />}
        path={identifier ? ROOT_PATH : "*"}
      />

      <Route element={<FullWidthLayout />}>{getLayoutRoutes("fullWidth")}</Route>

      <Route element={<FullWidthLayout background="bg-light" />}>
        {getLayoutRoutes("fullWidthGrey")}
      </Route>

      <Route element={<SiderLayout />}>{getLayoutRoutes("sider")}</Route>
    </Routes>
  );
}
