import axios from "axios";
import { z } from "zod";

import { Response, buildErrorResponse, buildSuccessResponse } from "src/adapters";
import {
  ID,
  IDParser,
  Message,
  Sorter,
  buildAuthorizationHeader,
  messageParser,
  serializeSorters,
} from "src/adapters/api";
import {
  datetimeParser,
  getListParser,
  getResourceParser,
  getStrictParser,
} from "src/adapters/parsers";
import {
  Credential,
  Env,
  IssuedQRCode,
  Json,
  Link,
  LinkStatus,
  ProofType,
  RefreshService,
} from "src/domain";
import { API_VERSION, QUERY_SEARCH_PARAM, STATUS_SEARCH_PARAM } from "src/utils/constants";
import { List, Resource } from "src/utils/types";

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

type CredentialInput = Omit<Credential, "proofTypes" | "createdAt" | "expiresAt"> & {
  createdAt: string;
  expiresAt: string | null;
  proofTypes: ProofTypeInput[];
};

export const credentialParser = getStrictParser<CredentialInput, Credential>()(
  z.object({
    createdAt: datetimeParser,
    credentialSubject: z.record(z.unknown()),
    expired: z.boolean(),
    expiresAt: datetimeParser.nullable(),
    id: z.string(),
    proofTypes: proofTypeParser,
    refreshService: z
      .object({ id: z.string(), type: z.literal("Iden3RefreshService2023") })
      .nullable(),
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
}): Promise<Response<Credential>> {
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
    return buildSuccessResponse(credentialParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function getCredentials({
  env,
  params: { did, maxResults, page, query, sorters, status },
  signal,
}: {
  env: Env;
  params: {
    did?: string;
    maxResults?: number;
    page?: number;
    query?: string;
    sorters?: Sorter[];
    status?: CredentialStatus;
  };
  signal?: AbortSignal;
}): Promise<Response<Resource<Credential>>> {
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
        ...(maxResults !== undefined ? { max_results: maxResults.toString() } : {}),
        ...(page !== undefined ? { page: page.toString() } : {}),
        ...(sorters !== undefined && sorters.length ? { sort: serializeSorters(sorters) } : {}),
      }),
      signal,
      url: `${API_VERSION}/credentials`,
    });
    return buildSuccessResponse(getResourceParser(credentialParser).parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export type CreateCredential = {
  credentialSchema: string;
  credentialSubject: Json;
  expiration: string | null;
  mtProof: boolean;
  refreshService: RefreshService | null;
  signatureProof: boolean;
  type: string;
};

export async function createCredential({
  env,
  payload,
}: {
  env: Env;
  payload: CreateCredential;
}): Promise<Response<ID>> {
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
    return buildSuccessResponse(IDParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function revokeCredential({
  env,
  nonce,
}: {
  env: Env;
  nonce: number;
}): Promise<Response<Message>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      url: `${API_VERSION}/credentials/revoke/${nonce}`,
    });
    return buildSuccessResponse(messageParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function deleteCredential({
  env,
  id,
}: {
  env: Env;
  id: string;
}): Promise<Response<Message>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "DELETE",
      url: `${API_VERSION}/credentials/${id}`,
    });
    return buildSuccessResponse(messageParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

// Links

export const linkStatusParser = getStrictParser<LinkStatus>()(
  z.union([z.literal("active"), z.literal("inactive"), z.literal("exceeded")])
);

type LinkInput = Omit<Link, "proofTypes" | "createdAt" | "credentialExpiration" | "expiration"> & {
  createdAt: string;
  credentialExpiration: string | null;
  expiration: string | null;
  proofTypes: ProofTypeInput[];
};

const linkParser = getStrictParser<LinkInput, Link>()(
  z.object({
    active: z.boolean(),
    createdAt: datetimeParser,
    credentialExpiration: datetimeParser.nullable(),
    credentialSubject: z.record(z.unknown()),
    expiration: datetimeParser.nullable(),
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
}): Promise<Response<Link>> {
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
    return buildSuccessResponse(linkParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
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
}): Promise<Response<List<Link>>> {
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
    return buildSuccessResponse(
      getListParser(linkParser)
        .transform(({ failed, successful }) => ({
          failed,
          successful: successful.sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime()),
        }))
        .parse(response.data)
    );
  } catch (error) {
    return buildErrorResponse(error);
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
}): Promise<Response<Message>> {
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
    return buildSuccessResponse(messageParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function deleteLink({
  env,
  id,
}: {
  env: Env;
  id: string;
}): Promise<Response<Message>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "DELETE",
      url: `${API_VERSION}/credentials/links/${id}`,
    });
    return buildSuccessResponse(messageParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export type CreateLink = {
  credentialExpiration: string | null;
  credentialSubject: Json;
  expiration: string | null;
  limitedClaims: number | null;
  mtProof: boolean;
  refreshService: RefreshService | null;
  schemaID: string;
  signatureProof: boolean;
};

export async function createLink({
  env,
  payload,
}: {
  env: Env;
  payload: CreateLink;
}): Promise<Response<ID>> {
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
    return buildSuccessResponse(IDParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

type AuthQRCodeInput = Omit<AuthQRCode, "linkDetail"> & {
  linkDetail: { proofTypes: ProofTypeInput[]; schemaType: string };
};

export type AuthQRCode = {
  linkDetail: { proofTypes: ProofType[]; schemaType: string };
  qrCodeLink: string;
  qrCodeRaw: string;
  sessionID: string;
};

const authQRCodeParser = getStrictParser<AuthQRCodeInput, AuthQRCode>()(
  z.object({
    linkDetail: z.object({ proofTypes: proofTypeParser, schemaType: z.string() }),
    qrCodeLink: z.string(),
    qrCodeRaw: z.string(),
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
}): Promise<Response<AuthQRCode>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      method: "POST",
      signal,
      url: `${API_VERSION}/credentials/links/${linkID}/qrcode`,
    });
    return buildSuccessResponse(authQRCodeParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

type IssuedQRCodeInput = {
  qrCodeLink: string;
  schemaType: string;
};

const issuedQRCodeParser = getStrictParser<IssuedQRCodeInput, IssuedQRCode>()(
  z
    .object({
      qrCodeLink: z.string(),
      schemaType: z.string(),
    })
    .transform(({ qrCodeLink, schemaType }) => ({ qrCode: qrCodeLink, schemaType: schemaType }))
);

export async function getIssuedQRCodes({
  credentialID,
  env,
  signal,
}: {
  credentialID: string;
  env: Env;
  signal: AbortSignal;
}): Promise<Response<[IssuedQRCode, IssuedQRCode]>> {
  try {
    const [qrLinkResponse, qrRawResponse] = await Promise.all([
      axios({
        baseURL: env.api.url,
        method: "GET",
        params: { type: "link" },
        signal,
        url: `${API_VERSION}/credentials/${credentialID}/qrcode`,
      }),
      axios({
        baseURL: env.api.url,
        method: "GET",
        params: { type: "raw" },
        signal,
        url: `${API_VERSION}/credentials/${credentialID}/qrcode`,
      }),
    ]);

    return buildSuccessResponse([
      issuedQRCodeParser.parse(qrLinkResponse.data),
      issuedQRCodeParser.parse(qrRawResponse.data),
    ]);
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export type ImportQRCode = {
  qrCode?: string;
  status: "done" | "pending" | "pendingPublish";
};

const importQRCodeParser = getStrictParser<ImportQRCode>()(
  z.object({
    qrCode: z.string().optional(),
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
}): Promise<Response<ImportQRCode>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      method: "GET",
      params: {
        sessionID,
      },
      url: `${API_VERSION}/credentials/links/${linkID}/qrcode`,
    });
    return buildSuccessResponse(importQRCodeParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}
