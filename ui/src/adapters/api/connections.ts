import axios from "axios";
import { z } from "zod";

import { RequestResponse } from "src/adapters";
import { Message, buildAuthorizationHeader, messageParser } from "src/adapters/api";
import { credentialParser } from "src/adapters/api/credentials";
import { getListParser, getStrictParser } from "src/adapters/parsers";
import { Connection, Env } from "src/domain";
import { API_VERSION, QUERY_SEARCH_PARAM } from "src/utils/constants";
import { buildAppError } from "src/utils/error";
import { List } from "src/utils/types";

type ConnectionInput = Omit<Connection, "credentials"> & {
  credentials: unknown[];
};

const connectionParser = getStrictParser<ConnectionInput, Connection>()(
  z.object({
    createdAt: z.coerce.date(z.string().datetime()),
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
}): Promise<RequestResponse<Connection>> {
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
    const data = connectionParser.parse(response.data);

    return { data, success: true };
  } catch (error) {
    return { error: buildAppError(error), success: false };
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
}): Promise<RequestResponse<List<Connection>>> {
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
    const data = getListParser(connectionParser).parse(response.data);

    return {
      data: {
        failed: data.failed,
        successful: data.successful.sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime()),
      },
      success: true,
    };
  } catch (error) {
    return { error: buildAppError(error), success: false };
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
}): Promise<RequestResponse<Message>> {
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

    const data = messageParser.parse(response.data);

    return { data, success: true };
  } catch (error) {
    return { error: buildAppError(error), success: false };
  }
}
