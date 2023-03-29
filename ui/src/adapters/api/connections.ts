import axios from "axios";
import { z } from "zod";

import { APIResponse, HTTPStatusSuccess, ResultOK, buildAPIError } from "src/adapters/api";
import { Env } from "src/domain";
import { buildAuthorizationHeader } from "src/utils/browser";
import { API_VERSION, QUERY_SEARCH_PARAM } from "src/utils/constants";
import { StrictSchema } from "src/utils/types";

//TODO move it to credentials when is properly cleaned
export interface Credential {
  attributes: {
    type: string;
  };
  id: string;
}

export interface Connection {
  createdAt: Date;
  credentials: Credential[];
  id: string;
  issuerID: string;
  userID: string;
}

export async function getConnection({
  env,
  id,
  signal,
}: {
  env: Env;
  id: string;
  signal: AbortSignal;
}): Promise<APIResponse<Connection>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "GET",
      signal,
      url: `${API_VERSION}/connections/${id}`,
    });
    const { data } = resultOKConnection.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
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
    createdAt: z.coerce.date(),
    credentials: z.array(credential),
    id: z.string(),
    issuerID: z.string(),
    userID: z.string(),
  })
);

export const resultOKConnection = StrictSchema<ResultOK<Connection>>()(
  z.object({
    data: connection,
    status: z.literal(HTTPStatusSuccess.OK),
  })
);

export const resultOKConnections = StrictSchema<ResultOK<Connection[]>>()(
  z.object({
    data: z.array(connection),
    status: z.literal(HTTPStatusSuccess.OK),
  })
);
