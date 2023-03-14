import { ComponentType } from "react";
import { Navigate, Route, Routes, generatePath } from "react-router-dom";

import { FullWidthLayout } from "src/components/layouts/FullWidthLayout";
import { SiderLayout } from "src/components/layouts/SiderLayout";
import { CreateSchema } from "src/components/schemas/CreateSchema";
import { ImportSchema } from "src/components/schemas/ImportSchema";
import { Issuance } from "src/components/schemas/Issuance";
import { ScanClaim } from "src/components/schemas/ScanClaim";
import { Schemas } from "src/components/schemas/Schemas";
import { NotFound } from "src/components/shared/NotFound";
import { Layout, ROUTES, RouteID } from "src/routes";
import { ROOT_PATH, SCHEMAS_TABS } from "src/utils/constants";

const COMPONENTS: Record<RouteID, ComponentType> = {
  claimLink: ScanClaim,
  createSchema: CreateSchema,
  importSchema: ImportSchema,
  issueClaim: Issuance,
  notFound: NotFound,
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
      <Route
        element={
          <Navigate
            to={generatePath(ROUTES.schemas.path, {
              tabID: SCHEMAS_TABS[0].tabID,
            })}
          />
        }
        path={ROOT_PATH}
      />
      <Route element={<FullWidthLayout />}>{getLayoutRoutes("fullWidth")}</Route>

      <Route element={<FullWidthLayout background="bg-light" />}>
        {getLayoutRoutes("fullWidthGrey")}
      </Route>

      <Route element={<SiderLayout />}>{getLayoutRoutes("sider")}</Route>
    </Routes>
  );
}
