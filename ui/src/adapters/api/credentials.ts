import axios from "axios";
import { z } from "zod";

import {
  APIResponse,
  GetAll,
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
import { Credential, Env, Json, Link, LinkStatus, ProofType } from "src/domain";
import { API_VERSION, QUERY_SEARCH_PARAM, STATUS_SEARCH_PARAM } from "src/utils/constants";

type ProofTypeInput = "BJJSignature2021" | "SparseMerkleTreeProof";

const proofTypeParser = getStrictParser<ProofTypeInput[], ProofType[]>()(
  z
    .array(z.union([z.literal("BJJSignature2021"), z.literal("SparseMerkleTreeProof")]))
    .min(1)
    .transform((values) =>
      values.map((value) => {
        switch (value) {
          case "BJJSignature2021": {
            return "SIG";
          }
          case "SparseMerkleTreeProof": {
            return "MTP";
          }
        }
      })
    )
);

// Credentials

export interface CredentialInput {
  createdAt: Date;
  credentialSubject: Record<string, unknown>;
  expired: boolean;
  expiresAt: Date | null;
  id: string;
  proofTypes: ProofTypeInput[];
  revNonce: number;
  revoked: boolean;
  schemaHash: string;
  schemaType: string;
  schemaUrl: string;
  userID: string;
}

export const credentialParser = getStrictParser<CredentialInput, Credential>()(
  z.object({
    createdAt: z.coerce.date(z.string().datetime()),
    credentialSubject: z.record(z.unknown()),
    expired: z.boolean(),
    expiresAt: z.coerce.date(z.string().datetime()).nullable(),
    id: z.string(),
    proofTypes: proofTypeParser,
    revNonce: z.number(),
    revoked: z.boolean(),
    schemaHash: z.string(),
    schemaType: z.string(),
    schemaUrl: z.string(),
    userID: z.string(),
  })
);

export type CredentialStatus = "all" | "revoked" | "expired";

export const credentialStatusParser = getStrictParser<CredentialStatus>()(
  z.union([z.literal("all"), z.literal("revoked"), z.literal("expired")])
);

export async function getCredential({
  credentialID,
  env,
  signal,
}: {
  credentialID: string;
  env: Env;
  signal?: AbortSignal;
}): Promise<APIResponse<Credential>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "GET",
      signal,
      url: `${API_VERSION}/credentials/${credentialID}`,
    });
    const { data } = resultOKCredentialParser.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

const resultOKCredentialParser = getStrictParser<ResultOK<CredentialInput>, ResultOK<Credential>>()(
  z.object({
    data: credentialParser,
    status: z.literal(200),
  })
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
}): Promise<APIResponse<GetAll<Credential>>> {
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
    const { data } = resultOKGetAllCredentialsParser.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

const resultOKGetAllCredentialsParser = getStrictParser<
  ResultOK<unknown[]>,
  ResultOK<GetAll<Credential>>
>()(
  z.object({
    data: z.array(z.unknown()).transform((unknowns) =>
      unknowns.reduce(
        (acc: GetAll<Credential>, curr: unknown, index): GetAll<Credential> => {
          const parsedCredential = credentialParser.safeParse(curr);

          return parsedCredential.success
            ? {
                ...acc,
                successful: [...acc.successful, parsedCredential.data],
              }
            : {
                ...acc,
                failed: [
                  ...acc.failed,
                  new z.ZodError<Credential>(
                    parsedCredential.error.issues.map((issue) => ({
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

// Links

export const linkStatusParser = getStrictParser<LinkStatus>()(
  z.union([z.literal("active"), z.literal("inactive"), z.literal("exceeded")])
);

interface LinkInput {
  active: boolean;
  createdAt: Date;
  credentialSubject: Record<string, unknown>;
  expiration: Date | null;
  id: string;
  issuedClaims: number;
  maxIssuance: number | null;
  proofTypes: ProofTypeInput[];
  schemaHash: string;
  schemaType: string;
  schemaUrl: string;
  status: LinkStatus;
}

const linkParser = getStrictParser<LinkInput, Link>()(
  z.object({
    active: z.boolean(),
    createdAt: z.coerce.date(z.string().datetime()),
    credentialSubject: z.record(z.unknown()),
    expiration: z.coerce.date(z.string().datetime()).nullable(),
    id: z.string(),
    issuedClaims: z.number(),
    maxIssuance: z.number().nullable(),
    proofTypes: proofTypeParser,
    schemaHash: z.string(),
    schemaType: z.string(),
    schemaUrl: z.string(),
    status: linkStatusParser,
  })
);

export async function getLink({
  env,
  linkID,
  signal,
}: {
  env: Env;
  linkID: string;
  signal: AbortSignal;
}): Promise<APIResponse<Link>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "GET",
      signal,
      url: `${API_VERSION}/credentials/links/${linkID}`,
    });
    const { data } = resultOKLinkParser.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

const resultOKLinkParser = getStrictParser<ResultOK<LinkInput>, ResultOK<Link>>()(
  z.object({
    data: linkParser,
    status: z.literal(200),
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
    const { data } = resultOKLinksParser.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

const resultOKLinksParser = getStrictParser<ResultOK<LinkInput[]>, ResultOK<Link[]>>()(
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
    const { data } = resultCreatedLinkParser.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

export interface AuthQRCodeInput {
  issuer: { displayName: string; logo: string };
  linkDetail: { proofTypes: ProofTypeInput[]; schemaType: string };
  qrCode?: unknown;
  sessionID: string;
}

export interface AuthQRCode {
  issuer: { displayName: string; logo: string };
  linkDetail: { proofTypes: ProofType[]; schemaType: string };
  qrCode?: unknown;
  sessionID: string;
}

const authQRCodeParser = getStrictParser<AuthQRCodeInput, AuthQRCode>()(
  z.object({
    issuer: z.object({
      displayName: z.string(),
      logo: z.string(),
    }),
    linkDetail: z.object({ proofTypes: proofTypeParser, schemaType: z.string() }),
    qrCode: z.unknown(),
    sessionID: z.string(),
  })
);

const resultOKAuthQRCodeParser = getStrictParser<ResultOK<AuthQRCodeInput>, ResultOK<AuthQRCode>>()(
  z.object({
    data: authQRCodeParser,
    status: z.literal(200),
  })
);

export async function createAuthQRCode({
  env,
  linkID,
  signal,
}: {
  env: Env;
  linkID: string;
  signal?: AbortSignal;
}): Promise<APIResponse<AuthQRCode>> {
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

    const { data } = resultOKAuthQRCodeParser.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

const resultCreatedLinkParser = getStrictParser<ResultCreated<ID>>()(
  z.object({
    data: IDParser,
    status: z.literal(201),
  })
);

export interface ImportQRCode {
  qrCode?: unknown;
  status: "done" | "pending" | "pendingPublish";
}

const importQRCheckDoneParser = getStrictParser<ImportQRCode>()(
  z.object({
    qrCode: z.unknown(),
    status: z.union([z.literal("done"), z.literal("pendingPublish"), z.literal("pending")]),
  })
);

const resultOKImportQRCheckParser = getStrictParser<ResultOK<ImportQRCode>>()(
  z.object({
    data: importQRCheckDoneParser,
    status: z.literal(200),
  })
);

export async function getImportQRCode({
  env,
  linkID,
  sessionID,
}: {
  env: Env;
  linkID: string;
  sessionID: string;
}): Promise<APIResponse<ImportQRCode>> {
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
    const { data } = resultOKImportQRCheckParser.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}
