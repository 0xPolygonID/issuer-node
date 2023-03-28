import axios from "axios";
import { z } from "zod";

import { Env } from "src/domain";
import { APIResponse, HTTPStatusSuccess, ResultOK, buildAPIError } from "src/utils/adapters";
import { QUERY_SEARCH_PARAM } from "src/utils/constants";
import { StrictSchema } from "src/utils/types";

const buildAuthorizationHeader = (env: Env) =>
  `Basic ${window.btoa(`${env.api.username}:${env.api.password}`)}`;

//TODO move it to credentials when is properly cleaned
export interface Credential {
  attributes: {
    type: string;
  };
  id: string;
}

export interface Connection {
  connection: {
    createdAt: string;
    id: string;
    issuerID: string;
    userID: string;
  };
  credentials?: Credential[];
}

export async function getConnections({
  credentials = true,
  env,
  params: { query },
  signal,
}: {
  credentials?: boolean;
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
      }),
      signal,
      url: `connections`,
    });
    const { data } = resultOKConnections.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

export const credential = StrictSchema<Credential>()(
  z.object({
    attributes: z.object({
      type: z.string(),
    }),
    id: z.string(),
  })
);

export const connection = StrictSchema<Connection>()(
  z.object({
    connection: z.object({
      createdAt: z.string(),
      id: z.string(),
      issuerID: z.string(),
      userID: z.string(),
    }),
    credentials: z.array(credential).optional(),
  })
);

export const resultOKConnections = StrictSchema<ResultOK<Connection[]>>()(
  z.object({
    data: z.array(connection),
    status: z.literal(HTTPStatusSuccess.OK),
  })
);
