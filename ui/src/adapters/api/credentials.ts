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
  CredentialProofType,
  Env,
  Identifier,
  IssuedQRCode,
  Json,
  Link,
  LinkStatus,
  ProofType,
  RefreshService,
} from "src/domain";
import { API_VERSION, QUERY_SEARCH_PARAM, STATUS_SEARCH_PARAM } from "src/utils/constants";
import { List, Resource } from "src/utils/types";

const proofTypeParser = getStrictParser<CredentialProofType[], ProofType[]>()(
  z
    .array(z.nativeEnum(CredentialProofType))
    .min(1)
    .transform((values) =>
      values.map((value) => {
        switch (value) {
          case CredentialProofType.BJJSignature2021: {
            return "SIG";
          }
          case CredentialProofType.Iden3SparseMerkleTreeProof: {
            return "MTP";
          }
        }
      })
    )
);

// Credentials

type CredentialInput = Pick<Credential, "id" | "revoked" | "schemaHash"> & {
  proofTypes: CredentialProofType[];
  vc: {
    credentialSchema: {
      id: string;
    } & Record<string, unknown>;
    credentialStatus: {
      revocationNonce: number;
    } & Record<string, unknown>;
    credentialSubject: {
      type: string;
    } & Record<string, unknown>;
    expirationDate?: string | null;
    issuanceDate: string;
    issuer: string;
    refreshService?: RefreshService | null;
  };
};

export const credentialParser = getStrictParser<CredentialInput, Credential>()(
  z
    .object({
      id: z.string(),
      proofTypes: proofTypeParser,
      revoked: z.boolean(),
      schemaHash: z.string(),
      vc: z.object({
        credentialSchema: z
          .object({
            id: z.string(),
          })
          .and(z.record(z.unknown())),
        credentialStatus: z
          .object({
            revocationNonce: z.number(),
          })
          .and(z.record(z.unknown())),
        credentialSubject: z
          .object({
            type: z.string(),
          })
          .and(z.record(z.unknown())),
        expirationDate: datetimeParser.nullable().default(null),
        issuanceDate: datetimeParser,
        issuer: z.string(),
        refreshService: z
          .object({ id: z.string(), type: z.literal("Iden3RefreshService2023") })
          .nullable()
          .default(null),
      }),
    })
    .transform(
      ({
        id,
        proofTypes,
        revoked,
        schemaHash,
        vc: {
          credentialSchema,
          credentialStatus,
          credentialSubject,
          expirationDate,
          issuanceDate,
          issuer,
          refreshService,
        },
      }) => {
        const expired = expirationDate ? new Date() > new Date(expirationDate) : false;

        return {
          createdAt: issuanceDate,
          credentialSubject,
          expired,
          expiresAt: expirationDate,
          id,
          proofTypes,
          refreshService: refreshService,
          revNonce: credentialStatus.revocationNonce,
          revoked,
          schemaHash,
          schemaType: credentialSubject.type,
          schemaUrl: credentialSchema.id,
          userID: issuer,
        };
      }
    )
);

export type CredentialStatus = "all" | "revoked" | "expired";

export const credentialStatusParser = getStrictParser<CredentialStatus>()(
  z.union([z.literal("all"), z.literal("revoked"), z.literal("expired")])
);

export async function getCredential({
  credentialID,
  env,
  identifier,
  signal,
}: {
  credentialID: string;
  env: Env;
  identifier: Identifier;
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
      url: `${API_VERSION}/identities/${identifier}/credentials/${credentialID}`,
    });
    return buildSuccessResponse(credentialParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function getCredentials({
  env,
  identifier,
  params: { credentialSubject, maxResults, page, query, sorters, status },
  signal,
}: {
  env: Env;
  identifier: Identifier;
  params: {
    credentialSubject?: string;
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
        ...(credentialSubject !== undefined ? { credentialSubject } : {}),
        ...(query !== undefined ? { [QUERY_SEARCH_PARAM]: query } : {}),
        ...(status !== undefined && status !== "all" ? { [STATUS_SEARCH_PARAM]: status } : {}),
        ...(maxResults !== undefined ? { max_results: maxResults.toString() } : {}),
        ...(page !== undefined ? { page: page.toString() } : {}),
        ...(sorters !== undefined && sorters.length ? { sort: serializeSorters(sorters) } : {}),
      }),
      signal,
      url: `${API_VERSION}/identities/${identifier}/credentials/search`,
    });
    return buildSuccessResponse(getResourceParser(credentialParser).parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export type CreateCredential = {
  credentialSchema: string;
  credentialSubject: Json;
  expiration: number | null;
  proofs: CredentialProofType[];
  refreshService: RefreshService | null;
  type: string;
};

export async function createCredential({
  env,
  identifier,
  payload,
}: {
  env: Env;
  identifier: Identifier;
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
      url: `${API_VERSION}/identities/${identifier}/credentials`,
    });
    return buildSuccessResponse(IDParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function revokeCredential({
  env,
  identifier,
  nonce,
}: {
  env: Env;
  identifier: Identifier;
  nonce: number;
}): Promise<Response<Message>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      url: `${API_VERSION}/identities/${identifier}/credentials/revoke/${nonce}`,
    });
    return buildSuccessResponse(messageParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function deleteCredential({
  env,
  id,
  identifier,
}: {
  env: Env;
  id: string;
  identifier: Identifier;
}): Promise<Response<Message>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "DELETE",
      url: `${API_VERSION}/identities/${identifier}/credentials/${id}`,
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
  proofTypes: CredentialProofType[];
};

const linkParser = getStrictParser<LinkInput, Link>()(
  z.object({
    active: z.boolean(),
    createdAt: datetimeParser,
    credentialExpiration: datetimeParser.nullable(),
    credentialSubject: z.record(z.unknown()),
    deepLink: z.string(),
    expiration: datetimeParser.nullable(),
    id: z.string(),
    issuedClaims: z.number(),
    maxIssuance: z.number().nullable(),
    proofTypes: proofTypeParser,
    schemaHash: z.string(),
    schemaType: z.string(),
    schemaUrl: z.string(),
    status: linkStatusParser,
    universalLink: z.string(),
  })
);

export async function getLink({
  env,
  identifier,
  linkID,
  signal,
}: {
  env: Env;
  identifier: Identifier;
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
      url: `${API_VERSION}/identities/${identifier}/credentials/links/${linkID}`,
    });
    return buildSuccessResponse(linkParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function getLinks({
  env,
  identifier,
  params: { query, status },
  signal,
}: {
  env: Env;
  identifier: Identifier;
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
      url: `${API_VERSION}/identities/${identifier}/credentials/links`,
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
  identifier,
  payload,
}: {
  env: Env;
  id: string;
  identifier: Identifier;
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
      url: `${API_VERSION}/identities/${identifier}/credentials/links/${id}`,
    });
    return buildSuccessResponse(messageParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function deleteLink({
  env,
  id,
  identifier,
}: {
  env: Env;
  id: string;
  identifier: Identifier;
}): Promise<Response<Message>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "DELETE",
      url: `${API_VERSION}/identities/${identifier}/credentials/links/${id}`,
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
  identifier,
  payload,
}: {
  env: Env;
  identifier: Identifier;
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
      url: `${API_VERSION}/identities/${identifier}/credentials/links`,
    });
    return buildSuccessResponse(IDParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

type AuthQRCodeInput = Omit<AuthQRCode, "linkDetail"> & {
  linkDetail: { proofTypes: CredentialProofType[]; schemaType: string };
};

export type AuthQRCode = {
  deepLink: string;
  linkDetail: { proofTypes: ProofType[]; schemaType: string };
  qrCodeRaw: string;
  universalLink: string;
};

const authQRCodeParser = getStrictParser<AuthQRCodeInput, AuthQRCode>()(
  z.object({
    deepLink: z.string(),
    linkDetail: z.object({ proofTypes: proofTypeParser, schemaType: z.string() }),
    qrCodeRaw: z.string(),
    universalLink: z.string(),
  })
);

export async function createAuthQRCode({
  env,
  identifier,
  linkID,
  signal,
}: {
  env: Env;
  identifier: Identifier;
  linkID: string;
  signal?: AbortSignal;
}): Promise<Response<AuthQRCode>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      method: "POST",
      signal,
      url: `${API_VERSION}/identities/${identifier}/credentials/links/${linkID}/qrcode`,
    });
    return buildSuccessResponse(authQRCodeParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

type IssuedQRCodeInput = {
  schemaType: string;
  universalLink: string;
};

const issuedQRCodeParser = getStrictParser<IssuedQRCodeInput, IssuedQRCode>()(
  z
    .object({
      schemaType: z.string(),
      universalLink: z.string(),
    })
    .transform(({ schemaType, universalLink }) => ({
      qrCode: universalLink,
      schemaType: schemaType,
    }))
);

export async function getIssuedQRCodes({
  credentialID,
  env,
  identifier,
  signal,
}: {
  credentialID: string;
  env: Env;
  identifier: Identifier;
  signal: AbortSignal;
}): Promise<Response<[IssuedQRCode, IssuedQRCode]>> {
  try {
    const [qrLinkResponse, qrRawResponse] = await Promise.all([
      axios({
        baseURL: env.api.url,
        headers: {
          Authorization: buildAuthorizationHeader(env),
        },
        method: "GET",
        params: { type: "deepLink" },
        signal,
        url: `${API_VERSION}/identities/${identifier}/credentials/${credentialID}/qrcode`,
      }),
      axios({
        baseURL: env.api.url,
        headers: {
          Authorization: buildAuthorizationHeader(env),
        },
        method: "GET",
        params: { type: "raw" },
        signal,
        url: `${API_VERSION}/identities/${identifier}/credentials/${credentialID}/qrcode`,
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
