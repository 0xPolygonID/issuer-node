import { App, ConfigProvider } from "antd";
import { extend as extendDayJsWith } from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";

import { Router } from "src/components/shared/Router";
import { RouterProvider } from "src/components/shared/RouterProvide";
import { EnvProvider } from "src/contexts/Env";
import { IdentityProvider } from "src/contexts/Identity";
import { IssuerStateProvider } from "src/contexts/IssuerState";
import { theme } from "src/styles/theme";
import { TOAST_NOTIFICATION_TIMEOUT } from "src/utils/constants";

import "src/styles/index.scss";

extendDayJsWith(relativeTime);

const rootElement = document.getElementById("root");

if (!rootElement) {
  throw new Error("Root HTML element could not be found in the DOM");
}

const root = createRoot(rootElement);

root.render(
  <StrictMode>
    <EnvProvider>
      <RouterProvider>
        <ConfigProvider theme={theme}>
          <App message={{ duration: TOAST_NOTIFICATION_TIMEOUT }}>
            <IdentityProvider>
              <IssuerStateProvider>
                <Router />
              </IssuerStateProvider>
            </IdentityProvider>
          </App>
        </ConfigProvider>
      </RouterProvider>
    </EnvProvider>
  </StrictMode>
);
