import axios from "axios";
import { z } from "zod";

import { Response, buildErrorResponse, buildSuccessResponse } from "src/adapters";
import { Message, buildAuthorizationHeader, messageParser } from "src/adapters/api";
import { credentialParser } from "src/adapters/api/credentials";
import {
  datetimeParser,
  getListParser,
  getResourceParser,
  getStrictParser,
} from "src/adapters/parsers";
import { Connection, Env } from "src/domain";
import { API_VERSION, QUERY_SEARCH_PARAM } from "src/utils/constants";
import { Resource } from "src/utils/types";

type ConnectionInput = Omit<Connection, "credentials" | "createdAt"> & {
  createdAt: string;
  credentials: unknown[];
};

const connectionParser = getStrictParser<ConnectionInput, Connection>()(
  z.object({
    createdAt: datetimeParser,
    credentials: getListParser(credentialParser),
    id: z.string(),
    issuerID: z.string(),
    userID: z.string(),
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
}): Promise<Response<Connection>> {
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
    return buildSuccessResponse(connectionParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function getConnections({
  credentials,
  env,
  params: { maxResults, page, query },
  signal,
}: {
  credentials: boolean;
  env: Env;
  params: {
    maxResults?: number;
    page?: number;
    query?: string;
  };
  signal?: AbortSignal;
}): Promise<Response<Resource<Connection>>> {
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
        ...(maxResults !== undefined ? { max_results: maxResults.toString() } : {}),
        ...(page !== undefined ? { page: page.toString() } : {}),
      }),
      signal,
      url: `${API_VERSION}/connections`,
    });
    return buildSuccessResponse(getResourceParser(connectionParser).parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
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
}): Promise<Response<Message>> {
  try {
    const response = await axios({
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
    return buildSuccessResponse(messageParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}
