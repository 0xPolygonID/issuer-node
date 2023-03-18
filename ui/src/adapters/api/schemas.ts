import axios from "axios";
import { z } from "zod";

import { JsonLdType } from "src/domain";
import { APIResponse, HTTPStatusSuccess, ResultOK, buildAPIError } from "src/utils/adapters";
import { API_PASSWORD, API_URL, API_USERNAME, QUERY_SEARCH_PARAM } from "src/utils/constants";
import { StrictSchema } from "src/utils/types";

export interface Schema {
  bigInt: string;
  createdAt: Date;
  hash: string;
  id: string;
  type: string;
  url: string;
}

export type SchemaAttribute = {
  description?: string;
  name: string;
  technicalName: string;
} & (
  | {
      type: "number" | "boolean" | "date";
    }
  | {
      type: "singlechoice";
      values: {
        name: string;
        value: number;
      }[];
    }
);

export async function importSchema({
  jsonLdType,
  url,
}: {
  jsonLdType: JsonLdType;
  url: string;
}): Promise<APIResponse<{ id: string }>> {
  try {
    const response = await axios({
      baseURL: API_URL,
      data: {
        schemaType: jsonLdType.name,
        url: url,
      },
      headers: {
        Authorization: `Basic ${window.btoa(`${API_USERNAME}:${API_PASSWORD}`)}`,
      },
      method: "POST",
      url: "schemas",
    });
    const { id } = z.object({ id: z.string() }).parse(response.data);

    return { data: { id }, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

export async function getSchema({
  schemaID,
  signal,
}: {
  schemaID: string;
  signal: AbortSignal;
}): Promise<APIResponse<Schema>> {
  try {
    const response = await axios({
      baseURL: API_URL,
      headers: {
        Authorization: `Basic ${window.btoa(`${API_USERNAME}:${API_PASSWORD}`)}`,
      },
      method: "GET",
      signal,
      url: `schemas/${schemaID}`,
    });
    const { data } = resultOKSchema.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

export async function getSchemas({
  params: { query },
  signal,
}: {
  params: {
    query?: string;
  };
  signal: AbortSignal;
}): Promise<
  APIResponse<{
    errors: z.ZodError<Schema>[];
    schemas: Schema[];
  }>
> {
  try {
    const response = await axios({
      baseURL: API_URL,
      headers: {
        Authorization: `Basic ${window.btoa(`${API_USERNAME}:${API_PASSWORD}`)}`,
      },
      method: "GET",
      params: new URLSearchParams({
        ...(query !== undefined ? { [QUERY_SEARCH_PARAM]: query } : {}),
      }),
      signal,
      url: "schemas",
    });
    const { data } = resultOKSchemasGetAll.parse(response);

    return {
      data: {
        errors: data.errors,
        schemas: data.schemas.sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime()),
      },
      isSuccessful: true,
    };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

export const schemaAttribute = StrictSchema<SchemaAttribute>()(
  z
    .object({
      description: z.string().optional(),
      name: z.string(),
      technicalName: z.string(),
    })
    .and(
      z.union([
        z.object({
          type: z.union([z.literal("boolean"), z.literal("number"), z.literal("date")]),
        }),
        z.object({
          type: z.literal("singlechoice"),
          values: z.array(
            z.object({
              name: z.string(),
              value: z.number(),
            })
          ),
        }),
      ])
    )
    .refine(
      (attribute) =>
        attribute.type !== "singlechoice" ||
        attribute.values.length ===
          new Set(attribute.values.map((singleChoice) => singleChoice.value)).size,
      (attribute) => ({
        message: `An error occurred validating the attribute "${attribute.name}". Single choice attributes can not have repeated numeric values`,
      })
    )
    .refine(
      (attribute) =>
        attribute.type !== "singlechoice" ||
        attribute.values.length ===
          new Set(attribute.values.map((singleChoice) => singleChoice.name)).size,
      (attribute) => ({
        message: `An error occurred validating the attribute "${attribute.name}". Single choice attributes can not have repeated values`,
      })
    )
);

export const schema = StrictSchema<Schema>()(
  z.object({
    bigInt: z.string(),
    createdAt: z.coerce.date(),
    hash: z.string(),
    id: z.string(),
    type: z.string(),
    url: z.string(),
  })
);

export const resultOKSchema = StrictSchema<ResultOK<Schema>>()(
  z.object({
    data: schema,
    status: z.literal(HTTPStatusSuccess.OK),
  })
);

interface SchemasGetAll {
  errors: z.ZodError<Schema>[];
  schemas: Schema[];
}

export const resultOKSchemasGetAll = StrictSchema<ResultOK<unknown[]>, ResultOK<SchemasGetAll>>()(
  z.object({
    data: z.array(z.unknown()).transform((unknowns) =>
      unknowns.reduce(
        (acc: SchemasGetAll, curr: unknown, index) => {
          const parsedSchema = schema.safeParse(curr);
          return parsedSchema.success
            ? {
                ...acc,
                schemas: [...acc.schemas, parsedSchema.data],
              }
            : {
                ...acc,
                errors: [
                  ...acc.errors,
                  new z.ZodError<Schema>(
                    parsedSchema.error.issues.map((issue) => ({
                      ...issue,
                      path: [index, ...issue.path],
                    }))
                  ),
                ],
              };
        },
        { errors: [], schemas: [] }
      )
    ),
    status: z.literal(HTTPStatusSuccess.OK),
  })
);
