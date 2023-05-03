import axios from "axios";
import { z } from "zod";

import { RequestResponse } from "src/adapters";
import { ID, IDParser, Message, buildAuthorizationHeader, messageParser } from "src/adapters/api";
import { getListParser, getStrictParser } from "src/adapters/parsers";
import { Credential, Env, IssuedQRCode, Json, Link, LinkStatus, ProofType } from "src/domain";
import { API_VERSION, QUERY_SEARCH_PARAM, STATUS_SEARCH_PARAM } from "src/utils/constants";
import { buildAppError } from "src/utils/error";
import { List } from "src/utils/types";

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

type CredentialInput = Omit<Credential, "proofTypes"> & {
  proofTypes: ProofTypeInput[];
};

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
}): Promise<RequestResponse<Credential>> {
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
    const data = credentialParser.parse(response.data);

    return { data, success: true };
  } catch (error) {
    return { error: buildAppError(error), success: false };
  }
}

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
}): Promise<RequestResponse<List<Credential>>> {
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
    const data = getListParser(credentialParser).parse(response.data);

    return {
      data: {
        failed: data.failed,
        successful: data.successful.sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime()),
      },
      success: true,
    };
  } catch (error) {
    return { error: buildAppError(error), success: false };
  }
}

export interface CreateCredential {
  credentialSchema: string;
  credentialSubject: Json;
  expiration: string | null;
  mtProof: boolean;
  signatureProof: boolean;
  type: string;
}

export async function createCredential({
  env,
  payload,
}: {
  env: Env;
  payload: CreateCredential;
}): Promise<RequestResponse<ID>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      data: payload,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      url: `${API_VERSION}/credentials`,
    });
    const data = IDParser.parse(response.data);

    return { data, success: true };
  } catch (error) {
    return { error: buildAppError(error), success: false };
  }
}

export async function revokeCredential({
  env,
  nonce,
}: {
  env: Env;
  nonce: number;
}): Promise<RequestResponse<Message>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      url: `${API_VERSION}/credentials/revoke/${nonce}`,
    });
    const data = messageParser.parse(response.data);

    return { data, success: true };
  } catch (error) {
    return { error: buildAppError(error), success: false };
  }
}

export async function deleteCredential({
  env,
  id,
}: {
  env: Env;
  id: string;
}): Promise<RequestResponse<Message>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "DELETE",
      url: `${API_VERSION}/credentials/${id}`,
    });

    const data = messageParser.parse(response.data);

    return { data, success: true };
  } catch (error) {
    return { error: buildAppError(error), success: false };
  }
}

// Links

export const linkStatusParser = getStrictParser<LinkStatus>()(
  z.union([z.literal("active"), z.literal("inactive"), z.literal("exceeded")])
);

type LinkInput = Omit<Link, "proofTypes"> & {
  proofTypes: ProofTypeInput[];
};

const linkParser = getStrictParser<LinkInput, Link>()(
  z.object({
    active: z.boolean(),
    createdAt: z.coerce.date(z.string().datetime()),
    credentialExpiration: z.coerce.date(z.string().datetime()).nullable(),
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
}): Promise<RequestResponse<Link>> {
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
    const data = linkParser.parse(response.data);

    return { data, success: true };
  } catch (error) {
    return { error: buildAppError(error), success: false };
  }
}

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
}): Promise<RequestResponse<List<Link>>> {
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
    const data = getListParser(linkParser).parse(response.data);

    return {
      data: {
        failed: data.failed,
        successful: data.successful.sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime()),
      },
      success: true,
    };
  } catch (error) {
    return { error: buildAppError(error), success: false };
  }
}

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
}): Promise<RequestResponse<Message>> {
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
    const data = messageParser.parse(response.data);

    return { data, success: true };
  } catch (error) {
    return { error: buildAppError(error), success: false };
  }
}

export async function deleteLink({
  env,
  id,
}: {
  env: Env;
  id: string;
}): Promise<RequestResponse<Message>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "DELETE",
      url: `${API_VERSION}/credentials/links/${id}`,
    });
    const data = messageParser.parse(response.data);

    return { data, success: true };
  } catch (error) {
    return { error: buildAppError(error), success: false };
  }
}

export interface CreateLink {
  credentialExpiration: string | null;
  credentialSubject: Json;
  expiration: string | null;
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
}): Promise<RequestResponse<ID>> {
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
    const data = IDParser.parse(response.data);

    return { data, success: true };
  } catch (error) {
    return { error: buildAppError(error), success: false };
  }
}

type AuthQRCodeInput = Omit<AuthQRCode, "linkDetail"> & {
  linkDetail: { proofTypes: ProofTypeInput[]; schemaType: string };
};

export interface AuthQRCode {
  linkDetail: { proofTypes: ProofType[]; schemaType: string };
  qrCode?: unknown;
  sessionID: string;
}

const authQRCodeParser = getStrictParser<AuthQRCodeInput, AuthQRCode>()(
  z.object({
    linkDetail: z.object({ proofTypes: proofTypeParser, schemaType: z.string() }),
    qrCode: z.unknown(),
    sessionID: z.string(),
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
}): Promise<RequestResponse<AuthQRCode>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      method: "POST",
      signal,
      url: `${API_VERSION}/credentials/links/${linkID}/qrcode`,
    });

    const data = authQRCodeParser.parse(response.data);

    return { data, success: true };
  } catch (error) {
    return { error: buildAppError(error), success: false };
  }
}

type IssuedQRCodeTypeInput = {
  body: {
    credentials: [
      {
        description: string;
      }
    ];
  };
};

const issuedQRCodeTypeParser = getStrictParser<IssuedQRCodeTypeInput, string>()(
  z
    .object({
      body: z.object({ credentials: z.tuple([z.object({ description: z.string() })]) }),
    })
    .transform((data) => data.body.credentials[0].description)
);

const issuedQRCodeParser = getStrictParser<unknown, IssuedQRCode>()(
  z.unknown().transform((unknown, context): IssuedQRCode => {
    const parsedSchemaType = issuedQRCodeTypeParser.safeParse(unknown);
    if (parsedSchemaType.success) {
      return {
        qrCode: unknown,
        schemaType: parsedSchemaType.data,
      };
    } else {
      parsedSchemaType.error.issues.map(context.addIssue);
      return z.NEVER;
    }
  })
);

export async function getIssuedQRCode({
  credentialID,
  env,
  signal,
}: {
  credentialID: string;
  env: Env;
  signal: AbortSignal;
}): Promise<RequestResponse<IssuedQRCode>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      method: "GET",
      signal,
      url: `${API_VERSION}/credentials/${credentialID}/qrcode`,
    });

    const data = issuedQRCodeParser.parse(response.data);

    return { data, success: true };
  } catch (error) {
    return { error: buildAppError(error), success: false };
  }
}

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

export async function getImportQRCode({
  env,
  linkID,
  sessionID,
}: {
  env: Env;
  linkID: string;
  sessionID: string;
}): Promise<RequestResponse<ImportQRCode>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      method: "GET",
      params: {
        sessionID,
      },
      url: `${API_VERSION}/credentials/links/${linkID}/qrcode`,
    });
    const data = importQRCheckDoneParser.parse(response.data);

    return { data, success: true };
  } catch (error) {
    return { error: buildAppError(error), success: false };
  }
}
