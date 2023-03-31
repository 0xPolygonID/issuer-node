import axios from "axios";
import { z } from "zod";

import {
  APIResponse,
  HTTPStatusSuccess,
  ResultOK,
  buildAPIError,
  buildAuthorizationHeader,
} from "src/adapters/api";
import { Credential, credential } from "src/adapters/api/credentials";
import { Env } from "src/domain";
import { API_VERSION, QUERY_SEARCH_PARAM } from "src/utils/constants";
import { StrictSchema } from "src/utils/types";

export interface Connection {
  createdAt: string;
  credentials: Credential[];
  id: string;
  issuerID: string;
  userID: string;
}

export async function getConnections({
  credentials,
  env,
  params: { query },
  signal,
}: {
  credentials: boolean;
  env: Env;
  params: {
    query?: string;
  };
  signal: AbortSignal;
}): Promise<APIResponse<Connection[]>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "GET",
      params: new URLSearchParams({
        ...(query !== undefined ? { [QUERY_SEARCH_PARAM]: query } : {}),
        ...(credentials ? { credentials: "true" } : {}),
      }),
      signal,
      url: `${API_VERSION}/connections`,
    });
    const { data } = resultOKConnections.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

export const connection = StrictSchema<Connection>()(
  z.object({
    createdAt: z.string(),
    credentials: z.array(credential),
    id: z.string(),
    issuerID: z.string(),
    userID: z.string(),
  })
);

export const resultOKConnections = StrictSchema<ResultOK<Connection[]>>()(
  z.object({
    data: z.array(connection),
    status: z.literal(HTTPStatusSuccess.OK),
  })
);
