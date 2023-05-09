import axios from "axios";
import { z } from "zod";

import { Response, buildErrorResponse, buildSuccessResponse } from "src/adapters";
import { ID, IDParser, buildAuthorizationHeader } from "src/adapters/api";
import { getListParser, getStrictParser } from "src/adapters/parsers";
import { Env, JsonLdType, Schema } from "src/domain";
import { API_VERSION, QUERY_SEARCH_PARAM } from "src/utils/constants";
import { List } from "src/utils/types";

const schemaParser = getStrictParser<Schema>()(
  z.object({
    bigInt: z.string(),
    createdAt: z.coerce.date(z.string().datetime()),
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

export async function getSchema({
  env,
  schemaID,
  signal,
}: {
  env: Env;
  schemaID: string;
  signal: AbortSignal;
}): Promise<Response<Schema>> {
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
    return buildSuccessResponse(schemaParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function getSchemas({
  env,
  params: { query },
  signal,
}: {
  env: Env;
  params: {
    query?: string;
  };
  signal: AbortSignal;
}): Promise<Response<List<Schema>>> {
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
      getListParser(schemaParser)
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
