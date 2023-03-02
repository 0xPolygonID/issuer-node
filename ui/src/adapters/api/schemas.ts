import axios from "axios";
import { z } from "zod";

import {
  APIResponse,
  HTTPStatusSuccess,
  ResultCreated,
  ResultOK,
  buildAPIError,
} from "src/utils/adapters";
import { ACTIVE_SEARCH_PARAM, API_URL, QUERY_SEARCH_PARAM } from "src/utils/constants";
import { StrictSchema } from "src/utils/types";

export interface PayloadSchemaCreate {
  attributes: SchemaAttribute[];
  mandatoryExpiration: boolean;
  schema: string;
  technicalName: string;
}

export type PayloadSchemaUpdate = {
  active: boolean;
};

export interface Schema {
  active: boolean;
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

export async function schemasCreate({
  id,
  payload,
  token,
}: {
  id: string;
  payload: PayloadSchemaCreate;
  token: string;
}): Promise<APIResponse<Schema>> {
  try {
    const response = await axios({
      baseURL: API_URL,
      data: payload,
      headers: {
        Authorization: `Bearer ${token}`,
      },
      method: "POST",
      url: `issuers/${id}/schemas`,
    });
    const { data } = resultCreatedSchema.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

export async function schemasGetSingle({
  issuerID,
  schemaID,
  signal,
  token,
}: {
  issuerID: string;
  schemaID: string;
  signal: AbortSignal;
  token: string;
}): Promise<APIResponse<Schema>> {
  try {
    const response = await axios({
      baseURL: API_URL,
      headers: {
        Authorization: `Bearer ${token}`,
      },
      method: "GET",
      signal,
      url: `issuers/${issuerID}/schemas/${schemaID}`,
    });
    const { data } = resultOKSchema.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

export async function schemasGetAll({
  id,
  params: { active, query },
  signal,
  token,
}: {
  id: string;
  params: {
    active?: boolean;
    query?: string;
  };
  signal: AbortSignal;
  token: string;
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
        Authorization: `Bearer ${token}`,
      },
      method: "GET",
      params: new URLSearchParams({
        ...(active !== undefined ? { [ACTIVE_SEARCH_PARAM]: active.toString() } : {}),
        ...(query !== undefined ? { [QUERY_SEARCH_PARAM]: query } : {}),
      }),
      signal,
      url: `issuers/${id}/schemas`,
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

export async function schemasUpdate({
  issuerID,
  payload,
  schemaID,
  token,
}: {
  issuerID: string;
  payload: PayloadSchemaUpdate;
  schemaID: string;
  token: string;
}): Promise<APIResponse<Schema>> {
  try {
    const response = await axios({
      baseURL: API_URL,
      data: payload,
      headers: {
        Authorization: `Bearer ${token}`,
      },
      method: "PATCH",
      url: `issuers/${issuerID}/schemas/${schemaID}`,
    });
    const { data } = resultOKSchema.parse(response);

    return { data, isSuccessful: true };
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

export const payloadSchemaCreate = StrictSchema<PayloadSchemaCreate>()(
  z.object({
    attributes: z.array(schemaAttribute),
    mandatoryExpiration: z.boolean(),
    schema: z.string(),
    technicalName: z.string(),
  })
);

export const schema = StrictSchema<Schema>()(
  z.object({
    active: z.boolean(),
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
