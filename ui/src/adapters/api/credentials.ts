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
import { getListParser, getStrictParser } from "src/adapters/parsers";
import { Credential, Env, IssuedQRCode, Json, Link, LinkStatus, ProofType } from "src/domain";
import { API_VERSION, QUERY_SEARCH_PARAM, STATUS_SEARCH_PARAM } from "src/utils/constants";
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
}): Promise<APIResponse<List<Credential>>> {
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
    const { data } = resultOKCredentialListParser.parse(response);

    return {
      data: {
        failed: data.failed,
        successful: data.successful.sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime()),
      },
      isSuccessful: true,
    };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

const resultOKCredentialListParser = getStrictParser<
  ResultOK<unknown[]>,
  ResultOK<List<Credential>>
>()(
  z.object({
    data: getListParser(credentialParser),
    status: z.literal(200),
  })
);

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
}): Promise<APIResponse<ID>> {
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
    const { data } = resultCreatedIDParser.parse(response);

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
}): Promise<APIResponse<List<Link>>> {
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
    const { data } = resultOKLinkListParser.parse(response);

    return {
      data: {
        failed: data.failed,
        successful: data.successful.sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime()),
      },
      isSuccessful: true,
    };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

const resultOKLinkListParser = getStrictParser<ResultOK<unknown[]>, ResultOK<List<Link>>>()(
  z.object({
    data: getListParser(linkParser),
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
    const { data } = resultCreatedIDParser.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
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

type IssuedQRCodeTypeInput = {
  body: {
    credentials: [
      {
        description: string;
      }
    ];
  };
};

const issuedQRCodeTypeParser = getStrictParser<IssuedQRCodeTypeInput, string | undefined>()(
  z
    .object({
      body: z.object({ credentials: z.tuple([z.object({ description: z.string() })]) }),
    })
    .transform((data) => data.body.credentials[0].description.split("#").pop())
);

type ResultOkIssuedQRCodeInput = {
  data?: unknown;
  status: 200;
};

const resultOKIssuedQRCodeParser = getStrictParser<
  ResultOkIssuedQRCodeInput,
  ResultOK<IssuedQRCode>
>()(
  z.object({
    data: z.unknown().transform((unknown): IssuedQRCode => {
      const parsedSchemaType = issuedQRCodeTypeParser.safeParse(unknown);
      const schemaType = parsedSchemaType.success ? parsedSchemaType.data : undefined;

      return {
        qrCode: unknown,
        schemaType,
      };
    }),
    status: z.literal(200),
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
}): Promise<APIResponse<IssuedQRCode>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      method: "GET",
      signal,
      url: `${API_VERSION}/credentials/${credentialID}/qrcode`,
    });

    const { data } = resultOKIssuedQRCodeParser.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

const resultCreatedIDParser = getStrictParser<ResultCreated<ID>>()(
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
