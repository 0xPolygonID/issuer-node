import { ComponentType } from "react";
import { Navigate, Route, Routes } from "react-router-dom";

import { ConnectionDetails } from "src/components/connections/ConnectionDetails";
import { ConnectionsTable } from "src/components/connections/ConnectionsTable";
import { CredentialDetails } from "src/components/credentials/CredentialDetails";
import { CredentialIssuedQR } from "src/components/credentials/CredentialIssuedQR";
import { Credentials } from "src/components/credentials/Credentials";
import { IssueCredential } from "src/components/credentials/IssueCredential";
import { LinkDetails } from "src/components/credentials/LinkDetails";
import { IssuerState } from "src/components/issuer-state/IssuerState";
import { CreateIssuer } from "src/components/issuers/CreateIssuer";
import { IssuerDetails } from "src/components/issuers/IssuerDetails";
import { Issuers } from "src/components/issuers/Issuers";
import { Onboarding } from "src/components/issuers/Onboarding";
import { FullWidthLayout } from "src/components/layouts/FullWidthLayout";
import { SiderLayout } from "src/components/layouts/SiderLayout";
import { ImportSchema } from "src/components/schemas/ImportSchema";
import { SchemaDetails } from "src/components/schemas/SchemaDetails";
import { Schemas } from "src/components/schemas/Schemas";
import { NotFound } from "src/components/shared/NotFound";
import { useIssuerContext } from "src/contexts/Issuer";
import { Layout, ROUTES, RouteID } from "src/routes";
import { ROOT_PATH } from "src/utils/constants";

const COMPONENTS: Record<RouteID, ComponentType> = {
  connectionDetails: ConnectionDetails,
  connections: ConnectionsTable,
  createIssuer: CreateIssuer,
  credentialDetails: CredentialDetails,
  credentialIssuedQR: CredentialIssuedQR,
  credentials: Credentials,
  importSchema: ImportSchema,
  issueCredential: IssueCredential,
  issuerDetails: IssuerDetails,
  issuers: Issuers,
  issuerState: IssuerState,
  linkDetails: LinkDetails,
  notFound: NotFound,
  onboarding: Onboarding,
  schemaDetails: SchemaDetails,
  schemas: Schemas,
};

export function Router() {
  const { issuerIdentifier } = useIssuerContext();

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
      {issuerIdentifier ? (
        <>
          <Route element={<Navigate to={ROUTES.schemas.path} />} path={ROOT_PATH} />

          <Route element={<FullWidthLayout />}>{getLayoutRoutes("fullWidth")}</Route>

          <Route element={<FullWidthLayout background="bg-light" />}>
            {getLayoutRoutes("fullWidthGrey")}
          </Route>

          <Route element={<SiderLayout />}>{getLayoutRoutes("sider")}</Route>
        </>
      ) : (
        <>
          <Route element={<FullWidthLayout />}>
            <Route element={<COMPONENTS.onboarding />} path={ROUTES.onboarding.path} />
          </Route>

          <Route element={<Navigate to={ROUTES.onboarding.path} />} path="*" />
        </>
      )}
    </Routes>
  );
}
