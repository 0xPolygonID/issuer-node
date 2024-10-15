import axios from "axios";
import { z } from "zod";

import { Response, buildErrorResponse, buildSuccessResponse } from "src/adapters";
import { ID, IDParser, buildAuthorizationHeader } from "src/adapters/api";
import { datetimeParser, getListParser, getStrictParser } from "src/adapters/parsers";
import { ApiSchema, Env, JsonLdType } from "src/domain";
import { getStorageByKey } from "src/utils/browser";
import { API_VERSION, IPFS_CUSTOM_GATEWAY_KEY, QUERY_SEARCH_PARAM } from "src/utils/constants";
import { List } from "src/utils/types";

type ApiSchemaInput = Omit<ApiSchema, "createdAt"> & {
  createdAt: string;
};

const apiSchemaParser = getStrictParser<ApiSchemaInput, ApiSchema>()(
  z.object({
    bigInt: z.string(),
    createdAt: datetimeParser,
    description: z.string().nullable(),
    hash: z.string(),
    id: z.string(),
    title: z.string().nullable(),
    type: z.string(),
    url: z.string(),
    version: z.string().nullable(),
  })
);

export async function importSchema({
  description,
  env,
  identifier,
  jsonLdType,
  schemaUrl,
  title,
  version,
}: {
  description?: string;
  env: Env;
  identifier: string;
  jsonLdType: JsonLdType;
  schemaUrl: string;
  title?: string;
  version?: string;
}): Promise<Response<ID>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      data: {
        description: description !== undefined ? description : null,
        schemaType: jsonLdType.name,
        title: title !== undefined ? title : null,
        url: schemaUrl,
        version: version !== undefined ? version : null,
      },
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      url: `${API_VERSION}/identities/${identifier}/schemas`,
    });
    return buildSuccessResponse(IDParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function getApiSchema({
  env,
  identifier,
  schemaID,
  signal,
}: {
  env: Env;
  identifier: string;
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
      url: `${API_VERSION}/identities/${identifier}/schemas/${schemaID}`,
    });
    return buildSuccessResponse(apiSchemaParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function getApiSchemas({
  env,
  identifier,
  params: { query },
  signal,
}: {
  env: Env;
  identifier: string;
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
      url: `${API_VERSION}/identities/${identifier}/schemas`,
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
  const ipfsGatewayUrl = getStorageByKey({
    defaultValue: env.ipfsGatewayUrl,
    key: IPFS_CUSTOM_GATEWAY_KEY,
    parser: z.string().url(),
  });
  const cid = ipfsUrl.split("ipfs://")[1];

  return cid !== undefined
    ? buildSuccessResponse(`${ipfsGatewayUrl}/ipfs/${cid}`)
    : buildErrorResponse("Invalid IPFS URL");
};

export const processUrl = (url: string, env: Env): Response<string> => {
  if (url.startsWith("ipfs://")) {
    return getIPFSGatewayUrl(env, url);
  } else {
    return { data: url, success: true };
  }
};
