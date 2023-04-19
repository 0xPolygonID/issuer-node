import { PropsWithChildren, createContext, useContext, useEffect, useMemo, useState } from "react";
import { z } from "zod";

import { EnvInput, envParser } from "src/adapters/env";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { Env } from "src/domain";
import { processZodError } from "src/utils/error";

const envDefaultValue: Env = {
  api: {
    password: "",
    url: "",
    username: "",
  },
  blockExplorerUrl: "",
  issuer: {
    did: "",
    name: "",
  },
};

const EnvContext = createContext(envDefaultValue);

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
    <ErrorResult
      error={[
        "An error occurred while reading the environment variables:\n",
        ...processZodError(value.error).map((e) => `"${e}"`),
        "\nPlease provide valid environment variables.",
      ].join("\n")}
    />
  );
}

export function useEnvContext() {
  return useContext(EnvContext);
}
