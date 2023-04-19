import { ConfigProvider, message } from "antd";
import { extend as extendDayJsWith } from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router-dom";

import { IssuerStateProvider } from "./contexts/IssuerState";
import { Router } from "src/components/shared/Router";
import { EnvProvider } from "src/contexts/Env";
import { theme } from "src/styles/theme";
import { TOAST_NOTIFICATION_TIMEOUT } from "src/utils/constants";

import "src/styles/index.scss";

extendDayJsWith(relativeTime);

const rootElement = document.getElementById("root");

if (!rootElement) {
  throw new Error("Root HTML element could not be found in the DOM");
}

const root = createRoot(rootElement);

message.config({ duration: TOAST_NOTIFICATION_TIMEOUT });

root.render(
  <StrictMode>
    <BrowserRouter>
      <ConfigProvider theme={theme}>
        <EnvProvider>
          <IssuerStateProvider>
            <Router />
          </IssuerStateProvider>
        </EnvProvider>
      </ConfigProvider>
    </BrowserRouter>
  </StrictMode>
);
