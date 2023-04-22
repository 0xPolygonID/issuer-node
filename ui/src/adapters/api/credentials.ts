import axios from "axios";
import { z } from "zod";

import {
  APIResponse,
  GetAll,
  HTTPStatusSuccess,
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

// Credentials

export const credentialParser = getStrictParser<Credential>()(
  z.object({
    createdAt: z.coerce.date(z.string().datetime()),
    credentialSubject: z.record(z.unknown()),
    expired: z.boolean(),
    expiresAt: z.coerce.date(z.string().datetime()).nullable(),
    id: z.string(),
    revNonce: z.number(),
    revoked: z.boolean(),
    schemaType: z.string(),
    schemaUrl: z.string(),
  })
);

export type CredentialStatus = "all" | "revoked" | "expired";

export const credentialStatusParser = getStrictParser<CredentialStatus>()(
  z.union([z.literal("all"), z.literal("revoked"), z.literal("expired")])
);

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
    status: z.literal(HTTPStatusSuccess.OK),
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

// Links

export const linkStatusParser = getStrictParser<LinkStatus>()(
  z.union([z.literal("active"), z.literal("inactive"), z.literal("exceeded")])
);

type ProofTypeInput = "BJJSignature2021" | "SparseMerkleTreeProof";

const proofTypeParser = getStrictParser<ProofTypeInput[], ProofType[]>()(
  z
    .array(z.union([z.literal("BJJSignature2021"), z.literal("SparseMerkleTreeProof")]))
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

interface LinkInput {
  active: boolean;
  credentialSubject: Record<string, unknown>;
  expiration: Date | null;
  id: string;
  issuedClaims: number;
  maxIssuance: number | null;
  proofTypes: ProofTypeInput[];
  schemaType: string;
  schemaUrl: string;
  status: LinkStatus;
}

const linkParser = getStrictParser<LinkInput, Link>()(
  z.object({
    active: z.boolean(),
    credentialSubject: z.record(z.unknown()),
    expiration: z.coerce.date(z.string().datetime()).nullable(),
    id: z.string(),
    issuedClaims: z.number(),
    maxIssuance: z.number().nullable(),
    proofTypes: proofTypeParser,
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
    status: z.literal(HTTPStatusSuccess.OK),
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

const resultCreatedLinkParser = getStrictParser<ResultCreated<ID>>()(
  z.object({
    data: IDParser,
    status: z.literal(HTTPStatusSuccess.Created),
  })
);

// QR codes

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

interface ShareCredentialQRCodeInput {
  issuer: { displayName: string; logo: string };
  linkDetail: LinkInput;
  qrCode: CredentialQRCode;
  sessionID: string;
}

export interface ShareCredentialQRCode {
  issuer: { displayName: string; logo: string };
  linkDetail: Link;
  qrCode: CredentialQRCode;
  sessionID: string;
}

const shareCredentialQRCodeParser = getStrictParser<
  ShareCredentialQRCodeInput,
  ShareCredentialQRCode
>()(
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

const resultOKShareCredentialQRCodeParser = getStrictParser<
  ResultOK<ShareCredentialQRCodeInput>,
  ResultOK<ShareCredentialQRCode>
>()(
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
