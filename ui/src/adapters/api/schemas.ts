import axios from "axios";
import { z } from "zod";

import {
  APIResponse,
  ID,
  IDParser,
  ResultOK,
  buildAPIError,
  buildAuthorizationHeader,
} from "src/adapters/api";
import { getStrictParser } from "src/adapters/parsers";
import { Env, JsonLdType, Schema } from "src/domain";
import { API_VERSION, QUERY_SEARCH_PARAM } from "src/utils/constants";
import { List } from "src/utils/types";

export async function importSchema({
  env,
  jsonLdType,
  schemaUrl,
}: {
  env: Env;
  jsonLdType: JsonLdType;
  schemaUrl: string;
}): Promise<APIResponse<ID>> {
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
    const { id } = IDParser.parse(response.data);

    return { data: { id }, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
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
}): Promise<APIResponse<Schema>> {
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
    const { data } = resultOKSchemaParser.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
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
}): Promise<APIResponse<List<Schema>>> {
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
    const { data } = resultOKSchemasParser.parse(response);

    return {
      data: {
        failed: data.failed,
        successful: data.successful.sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime()),
      },
      isSuccessful: true,
    };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

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

const resultOKSchemaParser = getStrictParser<ResultOK<Schema>>()(
  z.object({
    data: schemaParser,
    status: z.literal(200),
  })
);

const resultOKSchemasParser = getStrictParser<ResultOK<unknown[]>, ResultOK<List<Schema>>>()(
  z.object({
    data: z.array(z.unknown()).transform((unknowns) =>
      unknowns.reduce(
        (acc: List<Schema>, curr: unknown, index) => {
          const parsedSchema = schemaParser.safeParse(curr);

          return parsedSchema.success
            ? {
                ...acc,
                successful: [...acc.successful, parsedSchema.data],
              }
            : {
                ...acc,
                failed: [
                  ...acc.failed,
                  new z.ZodError<Schema>(
                    parsedSchema.error.issues.map((issue) => ({
                      ...issue,
                      path: [index, ...issue.path],
                    }))
                  ),
                ],
              };
        },
        { failed: [], successful: [] }
      )
    ),
    status: z.literal(200),
  })
);
