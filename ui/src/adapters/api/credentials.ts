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
import { schemaParser } from "src/adapters/api/schemas";
import { getStrictParser } from "src/adapters/parsers";
import { Credential, Env, Schema } from "src/domain";
import { API_VERSION, QUERY_SEARCH_PARAM } from "src/utils/constants";

export const credentialParser = getStrictParser<Credential>()(
  z.object({
    attributes: z.object({
      type: z.string(),
    }),
    id: z.string(),
  })
);

// TODO - refactor when credentials is implemented

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

interface CredentialInput {
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

export interface CredentialIssue {
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
  payload: CredentialIssue;
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
    const { data } = resultCreatedCredentialParser.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

export async function credentialUpdate({
  credentialID,
  env,
  payload,
  schemaID,
}: {
  credentialID: string;
  env: Env;
  payload: {
    active: boolean;
  };
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
    const { data } = resultOKCredentialParser.parse(response);

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
    const { data } = resultOKCredentialsGetAllParser.parse(response);

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

const credentialQRCodeParser = getStrictParser<CredentialQRCode>()(
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

interface ShareCredentialQRCodeInput {
  issuer: { displayName: string; logo: string };
  offerDetails: CredentialInput;
  qrcode: CredentialQRCode;
  sessionID: string;
}

const apiCredentialAttributeParser = getStrictParser<CredentialAttribute>()(
  z.object({
    attributeKey: z.string(),
    attributeValue: z.number(),
  })
);

const oldCredentialParser = getStrictParser<CredentialInput, OldCredential>()(
  z
    .object({
      active: z.boolean(),
      attributeValues: z.array(apiCredentialAttributeParser),
      claimLinkExpiration: z.coerce.date().nullable(),
      createdAt: z.coerce.date(),
      expiresAt: z.coerce.date().nullable(),
      id: z.string(),
      issuedClaims: z.number().nullable(),
      limitedClaims: z.number().nullable(),
      schemaTemplate: schemaParser,
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

const shareCredentialQRCodeParser = getStrictParser<
  ShareCredentialQRCodeInput,
  ShareCredentialQRCode
>()(
  z.object({
    issuer: z.object({
      displayName: z.string(),
      logo: z.string(),
    }),
    offerDetails: oldCredentialParser,
    qrcode: credentialQRCodeParser,
    sessionID: z.string(),
  })
);

const resultOKShareCredentialQRCodeParser = getStrictParser<
  ResultOK<ShareCredentialQRCodeInput>,
  ResultOK<ShareCredentialQRCode>
>()(
  z.object({
    data: shareCredentialQRCodeParser,
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

    const { data } = resultOKShareCredentialQRCodeParser.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

const resultCreatedCredentialParser = getStrictParser<
  ResultCreated<CredentialInput>,
  ResultCreated<OldCredential>
>()(
  z.object({
    data: oldCredentialParser,
    status: z.literal(HTTPStatusSuccess.Created),
  })
);

const resultOKCredentialParser = getStrictParser<
  ResultOK<CredentialInput>,
  ResultOK<OldCredential>
>()(
  z.object({
    data: oldCredentialParser,
    status: z.literal(HTTPStatusSuccess.OK),
  })
);

interface CredentialsGetAll {
  credentials: OldCredential[];
  errors: z.ZodError<OldCredential>[];
}

const resultOKCredentialsGetAllParser = getStrictParser<
  ResultOK<unknown[]>,
  ResultOK<CredentialsGetAll>
>()(
  z.object({
    data: z.array(z.unknown()).transform((unknowns) =>
      unknowns.reduce(
        (acc: CredentialsGetAll, curr: unknown, index) => {
          const parsedCredential = oldCredentialParser.safeParse(curr);

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

const addingQRCodeParser = getStrictParser<AddingQRCode>()(
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

const credentialQRCheckDoneParser = getStrictParser<CredentialQRCheckDone>()(
  z.object({
    qrcode: addingQRCodeParser,
    status: z.literal(CredentialQRStatus.Done),
  })
);

const credentialQRCheckErrorParser = getStrictParser<CredentialQRCheckError>()(
  z.object({
    status: z.literal(CredentialQRStatus.Error),
  })
);

const credentialQRCheckPendingParser = getStrictParser<CredentialQRCheckPending>()(
  z.object({
    status: z.literal(CredentialQRStatus.Pending),
  })
);

const credentialQRCheckParser = getStrictParser<CredentialQRCheck>()(
  z.union([
    credentialQRCheckDoneParser,
    credentialQRCheckErrorParser,
    credentialQRCheckPendingParser,
  ])
);

const resultOKCredentialQRCheck = getStrictParser<ResultOK<CredentialQRCheck>>()(
  z.object({
    data: credentialQRCheckParser,
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
