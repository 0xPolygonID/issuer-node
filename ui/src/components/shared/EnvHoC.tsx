import { PropsWithChildren } from "react";

import { envParser } from "src/adapters/env";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { Env } from "src/domain";
import { processZodError } from "src/utils/adapters";

const parsedEnv = envParser.safeParse(import.meta.env);
export const env: Env = parsedEnv.success
  ? parsedEnv.data
  : {
      api: {
        password: "",
        url: "",
        username: "",
      },
      issuer: {
        did: "",
        name: "",
      },
    };

export const apiAuth = `Basic ${env.api.username}:${env.api.password}`;

export function EnvHoC({ children }: PropsWithChildren) {
  return parsedEnv.success ? (
    <>{children}</>
  ) : (
    <ErrorResult
      error={[
        "An error occurred while reading the environment variables:\n",
        ...processZodError(parsedEnv.error).map((e) => `"${e}"`),
        "\nPlease provide valid environment variables.",
      ].join("\n")}
    />
  );
}
