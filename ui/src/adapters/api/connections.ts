import axios from "axios";
import { z } from "zod";

import { Response, buildErrorResponse, buildSuccessResponse } from "src/adapters";
import { Message, buildAuthorizationHeader, messageParser } from "src/adapters/api";
import { credentialParser } from "src/adapters/api/credentials";
import { datetimeParser, getListParser, getStrictParser } from "src/adapters/parsers";
import { Connection, Env } from "src/domain";
import { API_VERSION, QUERY_SEARCH_PARAM } from "src/utils/constants";
import { List } from "src/utils/types";

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
  params,
  signal,
}: {
  credentials: boolean;
  env: Env;
  params?: {
    query?: string;
  };
  signal?: AbortSignal;
}): Promise<Response<List<Connection>>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "GET",
      params: new URLSearchParams({
        ...(params?.query !== undefined ? { [QUERY_SEARCH_PARAM]: params?.query } : {}),
        ...(credentials ? { credentials: "true" } : {}),
      }),
      signal,
      url: `${API_VERSION}/connections`,
    });
    return buildSuccessResponse(
      getListParser(connectionParser)
        .transform(({ failed, successful }) => ({
          failed,
          successful: successful.sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime()),
        }))
        .parse(response.data)
    );
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
