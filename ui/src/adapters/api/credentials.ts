import axios from "axios";
import { z } from "zod";

import { Schema, schema } from "src/adapters/api/schemas";
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

export interface CredentialAttribute {
  attributeKey: string;
  attributeValue: number;
}

export interface Credential {
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

export interface CredentialPayload {
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
  claimLinkExpiration?: string;
  expirationDate?: string;
  limitedClaims?: number;
}

export async function credentialIssue({
  payload,
  schemaID,
}: {
  payload: CredentialIssuePayload;
  schemaID: string;
}): Promise<APIResponse<Credential>> {
  try {
    const response = await axios({
      baseURL: API_URL,
      data: payload,
      headers: {
        Authorization: `Basic ${API_USERNAME}:${API_PASSWORD}`,
      },
      method: "POST",
      url: `issuers/${ISSUER_DID}/schemas/${schemaID}/offers`,
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
  payload,
  schemaID,
}: {
  credentialID: string;
  payload: CredentialUpdatePayload;
  schemaID: string;
}): Promise<APIResponse<Credential>> {
  try {
    const response = await axios({
      baseURL: API_URL,
      data: payload,
      headers: {
        Authorization: `Basic ${API_USERNAME}:${API_PASSWORD}`,
      },
      method: "PATCH",
      url: `issuers/${ISSUER_DID}/schemas/${schemaID}/offers/${credentialID}`,
    });
    const { data } = resultOKCredential.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

export async function credentialsGetAll({
  params: { query, valid },
  signal,
}: {
  params: {
    query?: string;
    valid?: boolean;
  };
  signal?: AbortSignal;
}): Promise<
  APIResponse<{
    credentials: Credential[];
    errors: z.ZodError<Credential>[];
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
        ...(valid !== undefined ? { valid: valid.toString() } : {}),
      }),
      signal,
      url: `issuers/${ISSUER_DID}/offers`,
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
  offerDetails: Credential;
  qrcode: CredentialQRCode;
  sessionID: string;
}

export interface ShareCredentialQRCodePayload {
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

const credential = StrictSchema<CredentialPayload, Credential>()(
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
      }): Credential => ({
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

export const shareCredentialQRCode = StrictSchema<
  ShareCredentialQRCodePayload,
  ShareCredentialQRCode
>()(
  z.object({
    issuer: z.object({
      displayName: z.string(),
      logo: z.string(),
    }),
    offerDetails: credential,
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
  id,
  signal,
}: {
  id: string;
  signal?: AbortSignal;
}): Promise<APIResponse<ShareCredentialQRCode>> {
  try {
    const response = await axios({
      baseURL: API_URL,
      method: "POST",
      signal,
      url: `offers-qrcode/${id}`,
    });

    const { data } = resultOKShareCredentialQRCode.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

const resultCreatedCredential = StrictSchema<
  ResultCreated<CredentialPayload>,
  ResultCreated<Credential>
>()(
  z.object({
    data: credential,
    status: z.literal(HTTPStatusSuccess.Created),
  })
);

const resultOKCredential = StrictSchema<ResultOK<CredentialPayload>, ResultOK<Credential>>()(
  z.object({
    data: credential,
    status: z.literal(HTTPStatusSuccess.OK),
  })
);

interface CredentialsGetAll {
  credentials: Credential[];
  errors: z.ZodError<Credential>[];
}

const resultOKCredentialsGetAll = StrictSchema<ResultOK<unknown[]>, ResultOK<CredentialsGetAll>>()(
  z.object({
    data: z.array(z.unknown()).transform((unknowns) =>
      unknowns.reduce(
        (acc: CredentialsGetAll, curr: unknown, index) => {
          const parsedCredential = credential.safeParse(curr);
          return parsedCredential.success
            ? {
                ...acc,
                credentials: [...acc.credentials, parsedCredential.data],
              }
            : {
                ...acc,
                errors: [
                  ...acc.errors,
                  new z.ZodError<Credential>(
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

export const credentialQRCheck = StrictSchema<CredentialQRCheck>()(
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
  sessionID,
}: {
  credentialID: string;
  sessionID: string;
}): Promise<APIResponse<CredentialQRCheck>> {
  try {
    const response = await axios({
      baseURL: API_URL,
      method: "GET",
      params: {
        sessionID,
      },
      url: `offers-qrcode/${credentialID}`,
    });

    const { data } = resultOKCredentialQRCheck.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

export async function credentialsQRDownload({
  credentialID,
  sessionID,
}: {
  credentialID: string;
  sessionID: string;
}): Promise<APIResponse<Blob>> {
  try {
    const response = await axios({
      baseURL: API_URL,
      method: "GET",
      params: {
        sessionID,
      },
      responseType: "blob",
      url: `offers-qrcode/${credentialID}/download`,
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
