import axios from "axios";
import { z } from "zod";

import { Response, buildErrorResponse, buildSuccessResponse } from "src/adapters";
import { ID, IDParser, buildAuthorizationHeader } from "src/adapters/api";
import { datetimeParser, getListParser, getStrictParser } from "src/adapters/parsers";
import { ApiSchema, Env, JsonLdType } from "src/domain";
import { API_VERSION, QUERY_SEARCH_PARAM } from "src/utils/constants";
import { List } from "src/utils/types";

type ApiSchemaInput = Omit<ApiSchema, "createdAt"> & {
  createdAt: string;
};

const apiSchemaParser = getStrictParser<ApiSchemaInput, ApiSchema>()(
  z.object({
    bigInt: z.string(),
    createdAt: datetimeParser,
    hash: z.string(),
    id: z.string(),
    type: z.string(),
    url: z.string(),
  })
);

export async function importSchema({
  env,
  jsonLdType,
  schemaUrl,
}: {
  env: Env;
  jsonLdType: JsonLdType;
  schemaUrl: string;
}): Promise<Response<ID>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      data: {
        schemaType: jsonLdType.name,
        url: schemaUrl,
      },
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      url: `${API_VERSION}/schemas`,
    });
    return buildSuccessResponse(IDParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function getApiSchema({
  env,
  schemaID,
  signal,
}: {
  env: Env;
  schemaID: string;
  signal: AbortSignal;
}): Promise<Response<ApiSchema>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "GET",
      signal,
      url: `${API_VERSION}/schemas/${schemaID}`,
    });
    return buildSuccessResponse(apiSchemaParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function getApiSchemas({
  env,
  params: { query },
  signal,
}: {
  env: Env;
  params: {
    query?: string;
  };
  signal: AbortSignal;
}): Promise<Response<List<ApiSchema>>> {
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
      url: `${API_VERSION}/schemas`,
    });
    return buildSuccessResponse(
      getListParser(apiSchemaParser)
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

export const getIPFSGatewayUrl = (env: Env, ipfsUrl: string): Response<string> => {
  const cid = ipfsUrl.split("ipfs://")[1];

  return cid !== undefined
    ? buildSuccessResponse(`${env.ipfsGatewayUrl}/ipfs/${cid}`)
    : buildErrorResponse("Invalid IPFS URL");
};

export const processUrl = (url: string, env: Env): Response<string> => {
  if (url.startsWith("ipfs://")) {
    return getIPFSGatewayUrl(env, url);
  } else {
    return { data: url, success: true };
  }
};
