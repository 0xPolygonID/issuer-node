import axios from "axios";
import { z } from "zod";

import {
  APIResponse,
  HTTPStatusSuccess,
  ResultOK,
  buildAPIError,
  buildAuthorizationHeader,
} from "src/adapters/api";
import { credentialParser } from "src/adapters/api/credentials";
import { getStrictParser } from "src/adapters/parsers";
import { Connection, Env } from "src/domain";
import { API_VERSION, QUERY_SEARCH_PARAM } from "src/utils/constants";

const connectionParser = getStrictParser<Connection>()(
  z.object({
    createdAt: z.coerce.date(),
    credentials: z.array(credentialParser),
    id: z.string(),
    issuerID: z.string(),
    userID: z.string(),
  })
);

const resultOKConnectionParser = getStrictParser<ResultOK<Connection>>()(
  z.object({
    data: connectionParser,
    status: z.literal(HTTPStatusSuccess.OK),
  })
);

const resultOKConnectionsParser = getStrictParser<ResultOK<Connection[]>>()(
  z.object({
    data: z.array(connectionParser),
    status: z.literal(HTTPStatusSuccess.OK),
  })
);

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
    const { data } = resultOKConnectionParser.parse(response);

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
  signal?: AbortSignal;
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
    const { data } = resultOKConnectionsParser.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

export async function deleteConnection({
  deleteCredentials,
  env,
  id,
  revokeCredentials,
}: {
  deleteCredentials: boolean;
  env: Env;
  id: string;
  revokeCredentials: boolean;
}): Promise<APIResponse<string>> {
  try {
    const response = await axios<{ message: string }>({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "DELETE",
      params: new URLSearchParams({
        ...(revokeCredentials ? { revokeCredentials: "true" } : {}),
        ...(deleteCredentials ? { deleteCredentials: "true" } : {}),
      }),
      url: `${API_VERSION}/connections/${id}`,
    });

    return { data: response.data.message, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}
