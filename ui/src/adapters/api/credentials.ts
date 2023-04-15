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
import { getStrictParser } from "src/adapters/parsers";
import { Credential, Env, Json, Link, LinkAttribute, LinkStatus } from "src/domain";
import { API_VERSION, QUERY_SEARCH_PARAM, STATUS_SEARCH_PARAM } from "src/utils/constants";

export const credentialParser = getStrictParser<Credential>()(
  z.object({
    attributes: z.object({
      type: z.string(),
    }),
    createdAt: z.coerce.date(),
    expired: z.boolean().optional(),
    expiresAt: z.coerce.date().optional(),
    id: z.string(),
    revoked: z.boolean().optional(),
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
        ...(did !== undefined ? { did } : {}),
        ...(query !== undefined ? { [QUERY_SEARCH_PARAM]: query } : {}),
        ...(status !== undefined && status !== "all" ? { [STATUS_SEARCH_PARAM]: status } : {}),
      }),
      signal,
      url: `${API_VERSION}/credentials`,
    });
    const { data } = resultOKCredentialsParser.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

const resultOKCredentialsParser = getStrictParser<ResultOK<Credential[]>>()(
  z.object({
    data: z.array(credentialParser),
    status: z.literal(HTTPStatusSuccess.OK),
  })
);

export const linkStatusParser = getStrictParser<LinkStatus>()(
  z.union([z.literal("active"), z.literal("inactive"), z.literal("exceeded")])
);

const linkAttributeParser = getStrictParser<LinkAttribute>()(
  z.object({
    name: z.string(),
    value: z.string(),
  })
);

const linkParser = getStrictParser<Link>()(
  z.object({
    active: z.boolean(),
    attributes: z.array(linkAttributeParser),
    expiration: z.coerce.date().optional(),
    id: z.string(),
    issuedClaims: z.number(),
    maxIssuance: z.number().optional(),
    schemaType: z.string(),
    schemaUrl: z.string(),
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
  signal: AbortSignal;
}): Promise<APIResponse<Link[]>> {
  try {
    const response = await axios<Link[]>({
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
    const { data } = response;

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

export async function linkUpdate({
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
    const response = await axios<{ message: string }>({
      baseURL: env.api.url,
      data: payload,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "PATCH",
      url: `${API_VERSION}/credentials/links/${id}`,
    });

    return { data: response.data.message, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

export interface CreateLink {
  claimLinkExpiration: string | null;
  credentialSubject: Json;
  expirationDate: string | null;
  limitedClaims: number | null;
  mtProof: boolean;
  schemaID: string;
  signatureProof: boolean;
}

export async function createLink({
  env,
  payload,
}: {
  env: Env;
  payload: CreateLink;
}): Promise<APIResponse<{ id: string }>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      data: payload,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      url: `${API_VERSION}/credentials/links`,
    });
    const { data } = resultCreateLinkParser.parse(response);

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
    credentials: Credential[];
    errors: z.ZodError<Credential>[];
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
  linkDetail: Link;
  qrCode: CredentialQRCode;
  sessionID: string;
}

const shareCredentialQRCodeParser = getStrictParser<ShareCredentialQRCode>()(
  z.object({
    issuer: z.object({
      displayName: z.string(),
      logo: z.string(),
    }),
    linkDetail: linkParser,
    qrCode: credentialQRCodeParser,
    sessionID: z.string(),
  })
);

const resultOKShareCredentialQRCodeParser = getStrictParser<ResultOK<ShareCredentialQRCode>>()(
  z.object({
    data: shareCredentialQRCodeParser,
    status: z.literal(HTTPStatusSuccess.OK),
  })
);

export async function getCredentialLinkQRCode({
  env,
  linkID,
  signal,
}: {
  env: Env;
  linkID: string;
  signal?: AbortSignal;
}): Promise<APIResponse<ShareCredentialQRCode>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      signal,
      url: `${API_VERSION}/credentials/links/${linkID}/qrcode`,
    });

    const { data } = resultOKShareCredentialQRCodeParser.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

const resultCreateLinkParser = getStrictParser<ResultCreated<Link>>()(
  z.object({
    data: linkParser,
    status: z.literal(HTTPStatusSuccess.Created),
  })
);

interface CredentialsGetAll {
  credentials: Credential[];
  errors: z.ZodError<Credential>[];
}

const resultOKCredentialsGetAllParser = getStrictParser<
  ResultOK<unknown[]>,
  ResultOK<CredentialsGetAll>
>()(
  z.object({
    data: z.array(z.unknown()).transform((unknowns) =>
      unknowns.reduce(
        (acc: CredentialsGetAll, curr: unknown, index): CredentialsGetAll => {
          const parsedCredential = credentialParser.safeParse(curr);

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
  env,
  linkID,
  sessionID,
}: {
  env: Env;
  linkID: string;
  sessionID: string;
}): Promise<APIResponse<CredentialQRCheck>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      method: "GET",
      params: {
        sessionID,
      },
      url: `${API_VERSION}/offers-qrcode/${linkID}`,
    });

    const { data } = resultOKCredentialQRCheckParser.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}
