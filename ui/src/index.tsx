import { ConfigProvider, message } from "antd";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router-dom";

import { parsedEnv } from "src/adapters/parsers/env";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { Router } from "src/components/shared/Router";
import { theme } from "src/styles/theme";
import { processZodError } from "src/utils/adapters";
import { TOAST_NOTIFICATION_TIMEOUT } from "src/utils/constants";

import "src/styles/index.scss";

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
        {parsedEnv.success ? (
          <Router />
        ) : (
          <ErrorResult
            error={[
              "An error occurred while reading the environment variables:\n",
              ...processZodError(parsedEnv.error).map((e) => `"${e}"`),
              "\nPlease provide valid environment variables.",
            ].join("\n")}
          />
        )}
      </ConfigProvider>
    </BrowserRouter>
  </StrictMode>
);
