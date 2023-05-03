import { notification } from "antd";
import { ComponentType, useEffect, useState } from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import { z } from "zod";

import { ReactComponent as IconAlert } from "src/assets/icons/alert-triangle.svg";
import { ReactComponent as IconClose } from "src/assets/icons/x.svg";
import { ConnectionDetails } from "src/components/connections/ConnectionDetails";
import { ConnectionsTable } from "src/components/connections/ConnectionsTable";
import { CredentialDetails } from "src/components/credentials/CredentialDetails";
import { CredentialIssuedQR } from "src/components/credentials/CredentialIssuedQR";
import { CredentialLinkQR } from "src/components/credentials/CredentialLinkQR";
import { Credentials } from "src/components/credentials/Credentials";
import { IssueCredential } from "src/components/credentials/IssueCredential";
import { LinkDetails } from "src/components/credentials/LinkDetails";
import { IssuerState } from "src/components/issuer-state/IssuerState";
import { FullWidthLayout } from "src/components/layouts/FullWidthLayout";
import { SiderLayout } from "src/components/layouts/SiderLayout";
import { ImportSchema } from "src/components/schemas/ImportSchema";
import { SchemaDetails } from "src/components/schemas/SchemaDetails";
import { Schemas } from "src/components/schemas/Schemas";
import { NotFound } from "src/components/shared/NotFound";
import { useEnvContext } from "src/contexts/Env";
import { Layout, ROUTES, RouteID } from "src/routes";
import { ROOT_PATH } from "src/utils/constants";
import { getStorageByKey, setStorageByKey } from "src/utils/localStorage";

const COMPONENTS: Record<RouteID, ComponentType> = {
  connectionDetails: ConnectionDetails,
  connections: ConnectionsTable,
  credentialDetails: CredentialDetails,
  credentialIssuedQR: CredentialIssuedQR,
  credentialLinkQR: CredentialLinkQR,
  credentials: Credentials,
  importSchema: ImportSchema,
  issueCredential: IssueCredential,
  issuerState: IssuerState,
  linkDetails: LinkDetails,
  notFound: NotFound,
  schemaDetails: SchemaDetails,
  schemas: Schemas,
};

export function Router() {
  const warningKey = "warningNotification";

  const env = useEnvContext();
  const [isShowingWarning, setShowWarning] = useState(
    getStorageByKey({ defaultValue: true, key: warningKey, parser: z.boolean() })
  );

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

  useEffect(() => {
    if (env.warningMessage && isShowingWarning) {
      notification.warning({
        closeIcon: <IconClose />,
        description: env.warningMessage,
        duration: 0,
        icon: <IconAlert />,
        key: warningKey,
        message: "Warning",
        onClose: () => setShowWarning(setStorageByKey({ key: warningKey, value: false })),
        placement: "bottom",
      });
    }
  }, [env.warningMessage, isShowingWarning]);

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
