import axios from "axios";
import { z } from "zod";

import {
  APIResponse,
  HTTPStatusSuccess,
  ResultAccepted,
  ResultCreated,
  ResultOK,
  buildAPIError,
  buildAuthorizationHeader,
  resultOKMessage,
} from "src/adapters/api";
import { schemaParser } from "src/adapters/api/schemas";
import { getStrictParser } from "src/adapters/parsers";
import { Credential, Env, Link, LinkStatus, Schema } from "src/domain";
import { API_VERSION, QUERY_SEARCH_PARAM, STATUS_SEARCH_PARAM } from "src/utils/constants";

// TODO - refactor & order as Credentials are implemented

export const credentialParser = getStrictParser<Credential>()(
  z.object({
    createdAt: z.coerce.date(),
    credentialSubject: z.object({
      type: z.string(),
    }),
    expired: z.boolean(),
    expiresAt: z.coerce.date().optional(),
    id: z.string(),
    revNonce: z.number(),
    revoked: z.boolean(),
  })
);

export type CredentialStatus = "all" | "revoked" | "expired";

export const credentialStatusParser = getStrictParser<CredentialStatus>()(
  z.union([z.literal("all"), z.literal("revoked"), z.literal("expired")])
);

export async function getCredentials({
  env,
  params: { did, query, status },
  signal,
}: {
  env: Env;
  params: {
    did?: string;
    query?: string;
    status?: CredentialStatus;
  };
  signal?: AbortSignal;
}): Promise<APIResponse<Credential[]>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "GET",
      params: new URLSearchParams({
        ...(did !== undefined ? { did } : {}),
        ...(query !== undefined ? { [QUERY_SEARCH_PARAM]: query } : {}),
        ...(status !== undefined && status !== "all" ? { [STATUS_SEARCH_PARAM]: status } : {}),
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

const resultOKCredentials = getStrictParser<ResultOK<Credential[]>>()(
  z.object({
    data: z.array(credentialParser),
    status: z.literal(HTTPStatusSuccess.OK),
  })
);

export async function revokeCredential({
  env,
  nonce,
}: {
  env: Env;
  nonce: number;
}): Promise<APIResponse<string>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      url: `${API_VERSION}/credentials/revoke/${nonce}`,
    });
    const { data } = resultAcceptedMessage.parse(response);

    return { data: data.message, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

const resultAcceptedMessage = getStrictParser<ResultAccepted<{ message: string }>>()(
  z.object({
    data: z.object({
      message: z.string(),
    }),
    status: z.literal(HTTPStatusSuccess.Accepted),
  })
);

export async function deleteCredential({
  env,
  id,
}: {
  env: Env;
  id: string;
}): Promise<APIResponse<string>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "DELETE",
      url: `${API_VERSION}/credentials/${id}`,
    });

    const { data } = resultOKMessage.parse(response);

    return { data: data.message, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

export const linkStatusParser = getStrictParser<LinkStatus>()(
  z.union([z.literal("active"), z.literal("inactive"), z.literal("exceeded")])
);

export const link = getStrictParser<Link>()(
  z.object({
    active: z.boolean(),
    expiration: z.coerce.date().optional(),
    id: z.string(),
    issuedClaims: z.number(),
    maxIssuance: z.number().nullish(),
    schemaType: z.string(),
    status: linkStatusParser,
  })
);

export async function getLinks({
  env,
  params: { query, status },
  signal,
}: {
  env: Env;
  params: {
    query?: string;
    status?: LinkStatus;
  };
  signal?: AbortSignal;
}): Promise<APIResponse<Link[]>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "GET",
      params: new URLSearchParams({
        ...(query !== undefined ? { [QUERY_SEARCH_PARAM]: query } : {}),
        ...(status !== undefined ? { [STATUS_SEARCH_PARAM]: status } : {}),
      }),
      signal,
      url: `${API_VERSION}/credentials/links`,
    });
    const { data } = resultOKLinks.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

const resultOKLinks = getStrictParser<ResultOK<Link[]>>()(
  z.object({
    data: z.array(link),
    status: z.literal(HTTPStatusSuccess.OK),
  })
);

export async function updateLink({
  env,
  id,
  payload,
}: {
  env: Env;
  id: string;
  payload: {
    active: boolean;
  };
}): Promise<APIResponse<string>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      data: payload,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "PATCH",
      url: `${API_VERSION}/credentials/links/${id}`,
    });
    const { data } = resultOKMessage.parse(response);

    return { data: data.message, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

export async function deleteLink({
  env,
  id,
}: {
  env: Env;
  id: string;
}): Promise<APIResponse<string>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "DELETE",
      url: `${API_VERSION}/credentials/links/${id}`,
    });
    const { data } = resultOKMessage.parse(response);

    return { data: data.message, isSuccessful: true };
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

const resultOKCredentialQRCheckParser = getStrictParser<ResultOK<CredentialQRCheck>>()(
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

    const { data } = resultOKCredentialQRCheckParser.parse(response);

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
