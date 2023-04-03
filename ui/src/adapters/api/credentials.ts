import axios from "axios";
import { z } from "zod";

import {
  APIResponse,
  HTTPStatusSuccess,
  ResultCreated,
  ResultOK,
  buildAPIError,
  buildAuthorizationHeader,
} from "src/adapters/api";
import { Schema, schema } from "src/adapters/api/schemas";
import { Env } from "src/domain";
import { API_VERSION, QUERY_SEARCH_PARAM, TYPE_SEARCH_PARAM } from "src/utils/constants";
import { StrictSchema } from "src/utils/types";

export interface Credential {
  attributes: {
    type: string;
  };
  createdAt: Date;
  expiresAt?: Date;
  id: string;
  revoked?: boolean;
}

export const credential = StrictSchema<Credential>()(
  z.object({
    attributes: z.object({
      type: z.string(),
    }),
    createdAt: z.coerce.date(),
    expiresAt: z.optional(z.coerce.date()),
    id: z.string(),
    revoked: z.optional(z.boolean()),
  })
);

export type CredentialType = "all" | "revoked" | "expired";

export const typeParser = StrictSchema<CredentialType>()(
  z.union([z.literal("all"), z.literal("revoked"), z.literal("expired")])
);

export async function getCredentials({
  env,
  params: { query, type },
  signal,
}: {
  env: Env;
  params: {
    query?: string;
    type?: CredentialType;
  };
  signal: AbortSignal;
}): Promise<APIResponse<Credential[]>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "GET",
      params: new URLSearchParams({
        ...(query !== undefined ? { [QUERY_SEARCH_PARAM]: query } : {}),
        ...(type !== undefined && type !== "all" ? { [TYPE_SEARCH_PARAM]: type } : {}),
      }),
      signal,
      url: `${API_VERSION}/credentials`,
    });
    const { data } = resultOKCredentials.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

export const resultOKCredentials = StrictSchema<ResultOK<Credential[]>>()(
  z.object({
    data: z.array(credential),
    status: z.literal(HTTPStatusSuccess.OK),
  })
);

export interface CredentialAttribute {
  attributeKey: string;
  attributeValue: number;
}

export interface OldCredential {
  active: boolean;
  attributeValues: CredentialAttribute[];
  createdAt: Date;
  expiresAt: Date | null;
  id: string;
  linkAccessibleUntil: Date | null;
  linkCurrentIssuance: number | null;
  linkMaximumIssuance: number | null;
  schemaTemplate: Schema;
  valid: boolean;
}

interface CredentialPayload {
  active: boolean;
  attributeValues: CredentialAttribute[];
  claimLinkExpiration: Date | null;
  createdAt: Date;
  expiresAt: Date | null;
  id: string;
  issuedClaims: number | null;
  limitedClaims: number | null;
  schemaTemplate: Schema;
  valid: boolean;
}

export interface CredentialIssuePayload {
  attributes: CredentialAttribute[];
  claimLinkExpiration: string | null;
  expirationDate: string | null;
  limitedClaims: number | null;
}

export async function credentialIssue({
  env,
  payload,
  schemaID,
}: {
  env: Env;
  payload: CredentialIssuePayload;
  schemaID: string;
}): Promise<APIResponse<OldCredential>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      data: payload,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      url: `${API_VERSION}/issuers/${env.issuer.did}/schemas/${schemaID}/offers`,
    });
    const { data } = resultCreatedCredential.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

interface CredentialUpdatePayload {
  active: boolean;
}

export async function credentialUpdate({
  credentialID,
  env,
  payload,
  schemaID,
}: {
  credentialID: string;
  env: Env;
  payload: CredentialUpdatePayload;
  schemaID: string;
}): Promise<APIResponse<OldCredential>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      data: payload,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "PATCH",
      url: `${API_VERSION}/issuers/${env.issuer.did}/schemas/${schemaID}/offers/${credentialID}`,
    });
    const { data } = resultOKCredential.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

export async function credentialsGetAll({
  env,
  params: { query, valid },
  signal,
}: {
  env: Env;
  params: {
    query?: string;
    valid?: boolean;
  };
  signal?: AbortSignal;
}): Promise<
  APIResponse<{
    credentials: OldCredential[];
    errors: z.ZodError<OldCredential>[];
  }>
> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "GET",
      params: new URLSearchParams({
        ...(query !== undefined ? { [QUERY_SEARCH_PARAM]: query } : {}),
        ...(valid !== undefined ? { valid: valid.toString() } : {}),
      }),
      signal,
      url: `${API_VERSION}/issuers/${env.issuer.did}/offers`,
    });
    const { data } = resultOKCredentialsGetAll.parse(response);

    return {
      data: {
        credentials: data.credentials.sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime()),
        errors: data.errors,
      },
      isSuccessful: true,
    };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

interface CredentialQRCode {
  body: {
    callbackUrl: string;
    reason: string;
    scope: unknown[];
  };
  from: string;
  id: string;
  thid: string;
  typ: string;
  type: string;
}

const credentialQRCode = StrictSchema<CredentialQRCode>()(
  z.object({
    body: z.object({
      callbackUrl: z.string(),
      reason: z.string(),
      scope: z.array(z.unknown()),
    }),
    from: z.string(),
    id: z.string(),
    thid: z.string(),
    typ: z.string(),
    type: z.string(),
  })
);

export interface ShareCredentialQRCode {
  issuer: { displayName: string; logo: string };
  offerDetails: OldCredential;
  qrcode: CredentialQRCode;
  sessionID: string;
}

interface ShareCredentialQRCodePayload {
  issuer: { displayName: string; logo: string };
  offerDetails: CredentialPayload;
  qrcode: CredentialQRCode;
  sessionID: string;
}

const apiCredentialAttribute = StrictSchema<CredentialAttribute>()(
  z.object({
    attributeKey: z.string(),
    attributeValue: z.number(),
  })
);

const oldCredential = StrictSchema<CredentialPayload, OldCredential>()(
  z
    .object({
      active: z.boolean(),
      attributeValues: z.array(apiCredentialAttribute),
      claimLinkExpiration: z.coerce.date().nullable(),
      createdAt: z.coerce.date(),
      expiresAt: z.coerce.date().nullable(),
      id: z.string(),
      issuedClaims: z.number().nullable(),
      limitedClaims: z.number().nullable(),
      schemaTemplate: schema,
      valid: z.boolean(),
    })
    .transform(
      ({
        active,
        attributeValues,
        claimLinkExpiration: linkAccessibleUntil,
        createdAt,
        expiresAt,
        id,
        issuedClaims: linkCurrentIssuance,
        limitedClaims: linkMaximumIssuance,
        schemaTemplate,
        valid,
      }): OldCredential => ({
        active,
        attributeValues,
        createdAt,
        expiresAt,
        id,
        linkAccessibleUntil,
        linkCurrentIssuance,
        linkMaximumIssuance,
        schemaTemplate,
        valid,
      })
    )
);

const shareCredentialQRCode = StrictSchema<ShareCredentialQRCodePayload, ShareCredentialQRCode>()(
  z.object({
    issuer: z.object({
      displayName: z.string(),
      logo: z.string(),
    }),
    offerDetails: oldCredential,
    qrcode: credentialQRCode,
    sessionID: z.string(),
  })
);

const resultOKShareCredentialQRCode = StrictSchema<
  ResultOK<ShareCredentialQRCodePayload>,
  ResultOK<ShareCredentialQRCode>
>()(
  z.object({
    data: shareCredentialQRCode,
    status: z.literal(HTTPStatusSuccess.OK),
  })
);

export async function credentialsQRCreate({
  env,
  id,
  signal,
}: {
  env: Env;
  id: string;
  signal?: AbortSignal;
}): Promise<APIResponse<ShareCredentialQRCode>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      method: "POST",
      signal,
      url: `${API_VERSION}/offers-qrcode/${id}`,
    });

    const { data } = resultOKShareCredentialQRCode.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

const resultCreatedCredential = StrictSchema<
  ResultCreated<CredentialPayload>,
  ResultCreated<OldCredential>
>()(
  z.object({
    data: oldCredential,
    status: z.literal(HTTPStatusSuccess.Created),
  })
);

const resultOKCredential = StrictSchema<ResultOK<CredentialPayload>, ResultOK<OldCredential>>()(
  z.object({
    data: oldCredential,
    status: z.literal(HTTPStatusSuccess.OK),
  })
);

interface CredentialsGetAll {
  credentials: OldCredential[];
  errors: z.ZodError<OldCredential>[];
}

const resultOKCredentialsGetAll = StrictSchema<ResultOK<unknown[]>, ResultOK<CredentialsGetAll>>()(
  z.object({
    data: z.array(z.unknown()).transform((unknowns) =>
      unknowns.reduce(
        (acc: CredentialsGetAll, curr: unknown, index) => {
          const parsedCredential = oldCredential.safeParse(curr);
          return parsedCredential.success
            ? {
                ...acc,
                credentials: [...acc.credentials, parsedCredential.data],
              }
            : {
                ...acc,
                errors: [
                  ...acc.errors,
                  new z.ZodError<OldCredential>(
                    parsedCredential.error.issues.map((issue) => ({
                      ...issue,
                      path: [index, ...issue.path],
                    }))
                  ),
                ],
              };
        },
        { credentials: [], errors: [] }
      )
    ),
    status: z.literal(HTTPStatusSuccess.OK),
  })
);

interface AddingQRCode {
  body: {
    credentials: {
      description: string;
      id: string;
    }[];
    url: string;
  };
  from: string;
  id: string;
  thid: string;
  typ: string;
  type: string;
}

const addingQRCode = StrictSchema<AddingQRCode>()(
  z.object({
    body: z.object({
      credentials: z.array(
        z.object({
          description: z.string(),
          id: z.string(),
        })
      ),
      url: z.string(),
    }),
    from: z.string(),
    id: z.string(),
    thid: z.string(),
    typ: z.string(),
    type: z.string(),
  })
);

export enum CredentialQRStatus {
  Done = "done",
  Error = "error",
  Pending = "pending",
}

interface CredentialQRCheckDone {
  qrcode: AddingQRCode;
  status: CredentialQRStatus.Done;
}

interface CredentialQRCheckError {
  status: CredentialQRStatus.Error;
}

interface CredentialQRCheckPending {
  status: CredentialQRStatus.Pending;
}

export type CredentialQRCheck =
  | CredentialQRCheckDone
  | CredentialQRCheckError
  | CredentialQRCheckPending;

const credentialQRCheckDone = StrictSchema<CredentialQRCheckDone>()(
  z.object({
    qrcode: addingQRCode,
    status: z.literal(CredentialQRStatus.Done),
  })
);

const credentialQRCheckError = StrictSchema<CredentialQRCheckError>()(
  z.object({
    status: z.literal(CredentialQRStatus.Error),
  })
);

const credentialQRCheckPending = StrictSchema<CredentialQRCheckPending>()(
  z.object({
    status: z.literal(CredentialQRStatus.Pending),
  })
);

const credentialQRCheck = StrictSchema<CredentialQRCheck>()(
  z.union([credentialQRCheckDone, credentialQRCheckError, credentialQRCheckPending])
);

const resultOKCredentialQRCheck = StrictSchema<ResultOK<CredentialQRCheck>>()(
  z.object({
    data: credentialQRCheck,
    status: z.literal(HTTPStatusSuccess.OK),
  })
);

export async function credentialsQRCheck({
  credentialID,
  env,
  sessionID,
}: {
  credentialID: string;
  env: Env;
  sessionID: string;
}): Promise<APIResponse<CredentialQRCheck>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      method: "GET",
      params: {
        sessionID,
      },
      url: `${API_VERSION}/offers-qrcode/${credentialID}`,
    });

    const { data } = resultOKCredentialQRCheck.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

export async function credentialsQRDownload({
  credentialID,
  env,
  sessionID,
}: {
  credentialID: string;
  env: Env;
  sessionID: string;
}): Promise<APIResponse<Blob>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      method: "GET",
      params: {
        sessionID,
      },
      responseType: "blob",
      url: `${API_VERSION}/offers-qrcode/${credentialID}/download`,
    });

    if (response.data instanceof Blob) {
      return { data: response.data, isSuccessful: true };
    } else {
      return {
        error: { message: "Data returned by the API is not a valid file" },
        isSuccessful: false,
      };
    }
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}
