import { PropsWithChildren, createContext, useContext, useEffect, useMemo, useState } from "react";
import { z } from "zod";

import { EnvInput, envParser } from "src/adapters/env";
import { buildAppError, envErrorToString } from "src/adapters/parsers";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { Env } from "src/domain";

const defaultEnvContext: Env = {
  api: {
    password: "",
    url: "",
    username: "",
  },
  baseUrl: "",
  displayMethodBuilderUrl: "",
  ipfsGatewayUrl: "",
  issuer: {
    logo: "",
    name: "",
  },
  paymentPagesEnabled: false,
};

const EnvContext = createContext(defaultEnvContext);

export function EnvProvider(props: PropsWithChildren) {
  const [env, setEnv] = useState<z.SafeParseReturnType<EnvInput, Env>>();

  useEffect(() => {
    setEnv(envParser.safeParse(import.meta.env));
  }, []);

  const value = useMemo(() => {
    return env;
  }, [env]);

  return value?.success ? (
    <EnvContext.Provider value={value.data} {...props} />
  ) : !value ? null : (
    <ErrorResult error={envErrorToString(buildAppError(value.error))} />
  );
}

export function useEnvContext() {
  return useContext(EnvContext);
}
