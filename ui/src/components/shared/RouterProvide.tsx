import { PropsWithChildren } from "react";
import { BrowserRouter } from "react-router-dom";
import { useEnvContext } from "src/contexts/Env";
import { ROOT_PATH } from "src/utils/constants";

export function RouterProvider(props: PropsWithChildren) {
  const env = useEnvContext();
  const baseName = env.baseUrl || ROOT_PATH;

  return <BrowserRouter basename={baseName}>{props.children}</BrowserRouter>;
}
