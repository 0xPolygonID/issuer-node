import { FC, PropsWithChildren } from "react";

import { envParser } from "src/adapters/env";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { Env } from "src/domain";
import { processZodError } from "src/utils/adapters";

const envParsed = envParser.safeParse(import.meta.env);
export const env: Env = envParsed.success
  ? envParsed.data
  : {
      api: {
        password: "",
        url: "",
        username: "",
      },
      issuer: {
        did: "",
      },
    };

export const authorization = `Basic ${env.api.username}:${env.api.password}`;

const EnvHoC: FC<PropsWithChildren> = ({ children }) => {
  return envParsed.success ? (
    <>{children}</>
  ) : (
    <ErrorResult
      error={[
        "An error occurred while reading the environment variables:\n",
        ...processZodError(envParsed.error).map((e) => `"${e}"`),
        "\nPlease provide valid environment variables.",
      ].join("\n")}
    />
  );
};

export { EnvHoC };
