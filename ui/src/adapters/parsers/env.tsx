import { envParser } from "src/adapters/env";
import { Env } from "src/domain";

export const parsedEnv = envParser.safeParse(import.meta.env);

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
