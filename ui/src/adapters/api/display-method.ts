import axios from "axios";
import { z } from "zod";

import { Response, buildErrorResponse, buildSuccessResponse } from "src/adapters";
import {
  ID,
  IDParser,
  Message,
  Sorter,
  buildAuthorizationHeader,
  messageParser,
  serializeSorters,
} from "src/adapters/api";
import { getJsonFromUrl } from "src/adapters/json";
import { getResourceParser, getStrictParser } from "src/adapters/parsers";
import { DisplayMethod, DisplayMethodMetadata, DisplayMethodType, Env } from "src/domain";
import { API_VERSION } from "src/utils/constants";
import { Resource } from "src/utils/types";

export const displayMethodParser = getStrictParser<DisplayMethod>()(
  z.object({
    id: z.string(),
    name: z.string(),
    type: z.nativeEnum(DisplayMethodType),
    url: z.string().url(),
  })
);

export const displayMethodMetadataParser = getStrictParser<DisplayMethodMetadata>()(
  z.object({
    backgroundImageUrl: z.string().url(),
    description: z.string(),
    descriptionTextColor: z.string(),
    issuerName: z.string(),
    issuerTextColor: z.string(),
    logo: z.object({
      alt: z.string(),
      uri: z.string().url(),
    }),
    title: z.string(),
    titleTextColor: z.string(),
  })
);

export async function getDisplayMethods({
  env,
  identifier,
  params: { maxResults, page, sorters },
  signal,
}: {
  env: Env;
  identifier: string;
  params: {
    maxResults?: number;
    page?: number;
    sorters?: Sorter[];
  };
  signal?: AbortSignal;
}): Promise<Response<Resource<DisplayMethod>>> {
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
        ...(sorters !== undefined && sorters.length ? { sort: serializeSorters(sorters) } : {}),
      }),
      signal,
      url: `${API_VERSION}/identities/${identifier}/display-method`,
    });
    return buildSuccessResponse(getResourceParser(displayMethodParser).parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function getDisplayMethod({
  displayMethodID,
  env,
  identifier,
  signal,
}: {
  displayMethodID: string;
  env: Env;
  identifier: string;
  signal?: AbortSignal;
}): Promise<Response<DisplayMethod>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "GET",
      signal,
      url: `${API_VERSION}/identities/${identifier}/display-method/${displayMethodID}`,
    });
    return buildSuccessResponse(displayMethodParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function getDisplayMethodMetadata({
  env,
  signal,
  url,
}: {
  env: Env;
  signal?: AbortSignal;
  url: string;
}): Promise<Response<DisplayMethodMetadata>> {
  try {
    const json = await getJsonFromUrl({ env, signal, url });
    return buildSuccessResponse(displayMethodMetadataParser.parse(json));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export type UpsertDisplayMethod = {
  name: string;
  url: string;
};

export async function createDisplayMethod({
  env,
  identifier,
  payload,
}: {
  env: Env;
  identifier: string;
  payload: UpsertDisplayMethod;
}): Promise<Response<ID>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      data: payload,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      url: `${API_VERSION}/identities/${identifier}/display-method`,
    });
    return buildSuccessResponse(IDParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function deleteDisplayMethod({
  env,
  id,
  identifier,
}: {
  env: Env;
  id: string;
  identifier: string;
}): Promise<Response<Message>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "DELETE",
      url: `${API_VERSION}/identities/${identifier}/display-method/${id}`,
    });
    return buildSuccessResponse(messageParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function updateDisplayMethod({
  env,
  id,
  identifier,
  payload,
}: {
  env: Env;
  id: string;
  identifier: string;
  payload: UpsertDisplayMethod;
}): Promise<Response<Message>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      data: payload,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "PATCH",
      url: `${API_VERSION}/identities/${identifier}/display-method/${id}`,
    });
    return buildSuccessResponse(messageParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}
