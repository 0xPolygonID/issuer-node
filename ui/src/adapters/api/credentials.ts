import axios from "axios";
import { z } from "zod";

import {
  APIResponse,
  ID,
  IDParser,
  ResultAccepted,
  ResultCreated,
  ResultOK,
  buildAPIError,
  buildAuthorizationHeader,
  resultOKMessage,
} from "src/adapters/api";
import { getStrictParser } from "src/adapters/parsers";
import { Credential, Env, Json, Link, LinkStatus } from "src/domain";
import { API_VERSION, QUERY_SEARCH_PARAM, STATUS_SEARCH_PARAM } from "src/utils/constants";

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
    const { data } = resultOKCredentialsParser.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

const resultOKCredentialsParser = getStrictParser<ResultOK<Credential[]>>()(
  z.object({
    data: z.array(credentialParser),
    status: z.literal(200),
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
    status: z.literal(202),
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

const linkParser = getStrictParser<Link>()(
  z.object({
    active: z.boolean(),
    expiration: z.coerce.date().optional(),
    id: z.string(),
    issuedClaims: z.number(),
    maxIssuance: z.number().nullish(),
    proofTypes: z.array(
      z.union([z.literal("BJJSignature2021"), z.literal("SparseMerkleTreeProof")])
    ),
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
    data: z.array(linkParser),
    status: z.literal(200),
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

export interface ShareCredentialQRCode {
  issuer: { displayName: string; logo: string };
  linkDetail: Link;
  qrCode?: unknown;
  sessionID: string;
}

const shareCredentialQRCodeParser = getStrictParser<ShareCredentialQRCode>()(
  z.object({
    issuer: z.object({
      displayName: z.string(),
      logo: z.string(),
    }),
    linkDetail: linkParser,
    qrCode: z.unknown(),
    sessionID: z.string(),
  })
);

const resultOKShareCredentialQRCodeParser = getStrictParser<ResultOK<ShareCredentialQRCode>>()(
  z.object({
    data: shareCredentialQRCodeParser,
    status: z.literal(200),
  })
);

export async function createCredentialLinkQRCode({
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

const resultCreateLinkParser = getStrictParser<ResultCreated<ID>>()(
  z.object({
    data: IDParser,
    status: z.literal(201),
  })
);

export type CredentialQRStatus = "done" | "pending" | "pendingPublish";

export interface CredentialQRCheck {
  qrCode?: unknown;
  status: CredentialQRStatus;
}

const credentialQRCheckDoneParser = getStrictParser<CredentialQRCheck>()(
  z.object({
    qrCode: z.unknown(),
    status: z.union([z.literal("done"), z.literal("pendingPublish"), z.literal("pending")]),
  })
);

const resultOKCredentialQRCheckParser = getStrictParser<ResultOK<CredentialQRCheck>>()(
  z.object({
    data: credentialQRCheckDoneParser,
    status: z.literal(200),
  })
);

export async function getCredentialLinkQRCode({
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
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "GET",
      params: {
        sessionID,
      },
      url: `${API_VERSION}/credentials/links/${linkID}/qrcode`,
    });
    const { data } = resultOKCredentialQRCheckParser.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}
