import axios from "axios";
import { z } from "zod";

import {
  APIResponse,
  HTTPStatusSuccess,
  ResultCreated,
  ResultOK,
  buildAPIError,
} from "src/utils/adapters";
import {
  API_PASSWORD,
  API_URL,
  API_USERNAME,
  ISSUER_DID,
  QUERY_SEARCH_PARAM,
} from "src/utils/constants";
import { StrictSchema } from "src/utils/types";

export interface Schema {
  attributes: SchemaAttribute[];
  createdAt: Date;
  id: string;
  issuerID: string;
  mandatoryExpiration: boolean;
  schema: string;
  schemaHash: string;
  schemaURL: string;
  version: string;
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

export async function schemasGetSingle({
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
        Authorization: `Basic ${API_USERNAME}:${API_PASSWORD}`,
      },
      method: "GET",
      signal,
      url: `issuers/${ISSUER_DID}/schemas/${schemaID}`,
    });
    const { data } = resultOKSchema.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

export async function schemasGetAll({
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
        Authorization: `Basic ${API_USERNAME}:${API_PASSWORD}`,
      },
      method: "GET",
      params: new URLSearchParams({
        ...(query !== undefined ? { [QUERY_SEARCH_PARAM]: query } : {}),
      }),
      signal,
      url: `issuers/${ISSUER_DID}/schemas`,
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
    attributes: z.array(schemaAttribute),
    createdAt: z.coerce.date(),
    id: z.string(),
    issuerID: z.string(),
    mandatoryExpiration: z.boolean(),
    schema: z.string(),
    schemaHash: z.string(),
    schemaURL: z.string(),
    version: z.string(),
  })
);

export const resultCreatedSchema = StrictSchema<ResultCreated<Schema>>()(
  z.object({
    data: schema,
    status: z.literal(HTTPStatusSuccess.Created),
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
