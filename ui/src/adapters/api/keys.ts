import axios from "axios";
import { z } from "zod";

import { Response, buildErrorResponse, buildSuccessResponse } from "src/adapters";
import { ID, IDParser, Message, buildAuthorizationHeader, messageParser } from "src/adapters/api";
import { getResourceParser, getStrictParser } from "src/adapters/parsers";
import { Env, Key, KeyType } from "src/domain";
import { API_VERSION } from "src/utils/constants";
import { Resource } from "src/utils/types";

const keyParser = getStrictParser<Key>()(
  z.object({
    id: z.string(),
    isAuthCredential: z.boolean(),
    keyType: z.nativeEnum(KeyType),
    name: z.string(),
    publicKey: z.string(),
  })
);

export async function getKeys({
  env,
  identifier,
  params: { maxResults, page },
  signal,
}: {
  env: Env;
  identifier: string;
  params: {
    maxResults?: number;
    page?: number;
  };
  signal?: AbortSignal;
}): Promise<Response<Resource<Key>>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "GET",
      params: new URLSearchParams({
        ...(maxResults !== undefined ? { max_results: maxResults.toString() } : {}),
        ...(page !== undefined ? { page: page.toString() } : {}),
      }),
      signal,
      url: `${API_VERSION}/identities/${identifier}/keys`,
    });
    return buildSuccessResponse(getResourceParser(keyParser).parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function getKey({
  env,
  identifier,
  keyID,
  signal,
}: {
  env: Env;
  identifier: string;
  keyID: string;
  signal?: AbortSignal;
}): Promise<Response<Key>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "GET",
      signal,
      url: `${API_VERSION}/identities/${identifier}/keys/${keyID}`,
    });
    return buildSuccessResponse(keyParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export type CreateKey = {
  keyType: KeyType;
  name: string;
};

export async function createKey({
  env,
  identifier,
  payload,
}: {
  env: Env;
  identifier: string;
  payload: CreateKey;
}): Promise<Response<ID>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      data: payload,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      url: `${API_VERSION}/identities/${identifier}/keys`,
    });
    return buildSuccessResponse(IDParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export type UpdateKey = {
  name: string;
};

export async function updateKeyName({
  env,
  identifier,
  keyID,
  payload,
}: {
  env: Env;
  identifier: string;
  keyID: string;
  payload: UpdateKey;
}) {
  try {
    await axios({
      baseURL: env.api.url,
      data: payload,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "PATCH",
      url: `${API_VERSION}/identities/${identifier}/keys/${keyID}`,
    });

    return buildSuccessResponse(undefined);
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function deleteKey({
  env,
  identifier,
  keyID,
}: {
  env: Env;
  identifier: string;
  keyID: string;
}): Promise<Response<Message>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "DELETE",
      url: `${API_VERSION}/identities/${identifier}/keys/${keyID}`,
    });
    return buildSuccessResponse(messageParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}
