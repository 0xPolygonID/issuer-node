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
import { datetimeParser, getResourceParser, getStrictParser } from "src/adapters/parsers";
import {
  Credential,
  DisplayMethod,
  Env,
  IssuedMessage,
  Json,
  Link,
  LinkStatus,
  ProofType,
  RefreshService,
} from "src/domain";
import { API_VERSION, QUERY_SEARCH_PARAM, STATUS_SEARCH_PARAM } from "src/utils/constants";
import { Resource } from "src/utils/types";

// Credentials

type CredentialInput = Pick<Credential, "id" | "revoked" | "schemaHash"> & {
  proofTypes: ProofType[];
  vc: {
    credentialSchema: {
      id: string;
    } & Record<string, unknown>;
    credentialStatus: {
      revocationNonce: number;
    } & Record<string, unknown>;
    credentialSubject: Record<string, unknown>;
    displayMethod?: DisplayMethod | null;
    expirationDate?: string | null;
    issuanceDate: string;
    issuer: string;
    refreshService?: RefreshService | null;
    type: [string, string];
  };
};

export const credentialParser = getStrictParser<CredentialInput, Credential>()(
  z
    .object({
      id: z.string(),
      proofTypes: z.array(z.nativeEnum(ProofType)),
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
        credentialSubject: z.record(z.unknown()),
        displayMethod: z
          .object({ id: z.string(), type: z.literal("Iden3BasicDisplayMethodv2") })
          .nullable()
          .default(null),
        expirationDate: datetimeParser.nullable().default(null),
        issuanceDate: datetimeParser,
        issuer: z.string(),
        refreshService: z
          .object({ id: z.string(), type: z.literal("Iden3RefreshService2023") })
          .nullable()
          .default(null),
        type: z.tuple([z.string(), z.string()]),
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
          displayMethod,
          expirationDate,
          issuanceDate,
          issuer,
          refreshService,
          type,
        },
      }) => {
        const expired = expirationDate ? new Date() > new Date(expirationDate) : false;
        const [, schemaType] = type;

        return {
          credentialSubject,
          displayMethod,
          expirationDate,
          expired,
          id,
          issuanceDate,
          proofTypes,
          refreshService,
          revNonce: credentialStatus.revocationNonce,
          revoked,
          schemaHash,
          schemaType,
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
  identifier: string;
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
  identifier: string;
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
      url: `${API_VERSION}/identities/${identifier}/credentials`,
    });
    return buildSuccessResponse(getResourceParser(credentialParser).parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export type CreateCredential = {
  credentialSchema: string;
  credentialSubject: Json;
  displayMethod: DisplayMethod | null;
  expiration: number | null;
  proofs: ProofType[];
  refreshService: RefreshService | null;
  type: string;
};

export async function createCredential({
  env,
  identifier,
  payload,
}: {
  env: Env;
  identifier: string;
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
  identifier: string;
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
  identifier: string;
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
  proofTypes: ProofType[];
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
    proofTypes: z.array(z.nativeEnum(ProofType)),
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
  identifier: string;
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
  params: { maxResults, page, query, sorters, status },
  signal,
}: {
  env: Env;
  identifier: string;
  params: {
    maxResults?: number;
    page?: number;
    query?: string;
    sorters?: Sorter[];
    status?: LinkStatus;
  };
  signal?: AbortSignal;
}): Promise<Response<Resource<Link>>> {
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
        ...(maxResults !== undefined ? { max_results: maxResults.toString() } : {}),
        ...(page !== undefined ? { page: page.toString() } : {}),
        ...(sorters !== undefined && sorters.length ? { sort: serializeSorters(sorters) } : {}),
      }),
      signal,
      url: `${API_VERSION}/identities/${identifier}/credentials/links`,
    });
    return buildSuccessResponse(getResourceParser(linkParser).parse(response.data));
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
  identifier: string;
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
  identifier: string;
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
  displayMethod: DisplayMethod | null;
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
  identifier: string;
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

export type AuthRequestMessage = {
  deepLink: string;
  universalLink: string;
};

const authRequestMessageParser = getStrictParser<AuthRequestMessage>()(
  z.object({
    deepLink: z.string(),
    universalLink: z.string(),
  })
);

export async function createAuthRequestMessage({
  env,
  identifier,
  linkID,
  signal,
}: {
  env: Env;
  identifier: string;
  linkID: string;
  signal?: AbortSignal;
}): Promise<Response<AuthRequestMessage>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      method: "POST",
      signal,
      url: `${API_VERSION}/identities/${identifier}/credentials/links/${linkID}/offer`,
    });
    return buildSuccessResponse(authRequestMessageParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

const issuedMessageParser = getStrictParser<IssuedMessage>()(
  z.object({
    schemaType: z.string(),
    universalLink: z.string(),
  })
);

export async function getIssuedMessages({
  credentialID,
  env,
  identifier,
  signal,
}: {
  credentialID: string;
  env: Env;
  identifier: string;
  signal?: AbortSignal;
}): Promise<Response<[IssuedMessage, IssuedMessage]>> {
  try {
    const [universalLinkResponse, deepLinkResponse] = await Promise.all([
      axios({
        baseURL: env.api.url,
        headers: {
          Authorization: buildAuthorizationHeader(env),
        },
        method: "GET",
        signal,
        url: `${API_VERSION}/identities/${identifier}/credentials/${credentialID}/offer`,
      }),
      axios({
        baseURL: env.api.url,
        headers: {
          Authorization: buildAuthorizationHeader(env),
        },
        method: "GET",
        params: { type: "deepLink" },
        signal,
        url: `${API_VERSION}/identities/${identifier}/credentials/${credentialID}/offer`,
      }),
    ]);

    return buildSuccessResponse([
      issuedMessageParser.parse(universalLinkResponse.data),
      issuedMessageParser.parse(deepLinkResponse.data),
    ]);
  } catch (error) {
    return buildErrorResponse(error);
  }
}
