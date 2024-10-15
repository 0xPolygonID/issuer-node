import axios from "axios";
import { z } from "zod";

import { Response, buildErrorResponse, buildSuccessResponse } from "src/adapters";
import { buildAuthorizationHeader } from "src/adapters/api";
import { getListParser, getStrictParser } from "src/adapters/parsers";
import {
  CredentialStatusType,
  Env,
  Identity,
  IdentityDetails,
  IdentityType,
  Method,
  SupportedNetwork,
} from "src/domain";

import { API_VERSION } from "src/utils/constants";
import { List } from "src/utils/types";

const identityParser = getStrictParser<Identity>()(
  z.object({
    blockchain: z.string(),
    credentialStatusType: z.nativeEnum(CredentialStatusType),
    displayName: z.string().nullable(),
    identifier: z.string(),
    method: z.nativeEnum(Method),
    network: z.string(),
  })
);

export const identifierParser = getStrictParser<string>()(z.string());

export async function getIdentities({
  env,
  signal,
}: {
  env: Env;
  signal: AbortSignal;
}): Promise<Response<List<Identity>>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "GET",
      signal,
      url: `${API_VERSION}/identities`,
    });

    return buildSuccessResponse(getListParser(identityParser).parse(response.data || []));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export type CreateIdentity = {
  blockchain: string;
  credentialStatusType: CredentialStatusType;
  displayName: string;
  method: string;
  network: string;
  type: IdentityType;
};

type CreateIdentityResponse = {
  identifier: string;
};

export const createIdentityResponseParser = getStrictParser<CreateIdentityResponse>()(
  z.object({
    identifier: z.string(),
  })
);

export async function createIdentity({
  env,
  payload,
}: {
  env: Env;
  payload: CreateIdentity;
}): Promise<Response<CreateIdentityResponse>> {
  try {
    const { credentialStatusType, displayName, ...didMetadata } = payload;
    const response = await axios({
      baseURL: env.api.url,
      data: { credentialStatusType, didMetadata, displayName },
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      url: `${API_VERSION}/identities`,
    });

    return buildSuccessResponse(createIdentityResponseParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export const identityDetailsParser = getStrictParser<IdentityDetails>()(
  z.object({
    credentialStatusType: z.nativeEnum(CredentialStatusType),
    displayName: z.string().nullable(),
    identifier: z.string(),
    keyType: z.nativeEnum(IdentityType),
  })
);

export async function getIdentity({
  env,
  identifier,
  signal,
}: {
  env: Env;
  identifier: string;
  signal?: AbortSignal;
}): Promise<Response<IdentityDetails>> {
  try {
    const response = await axios({
      baseURL: env.api.url,

      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "GET",
      signal,
      url: `${API_VERSION}/identities/${identifier}`,
    });

    return buildSuccessResponse(identityDetailsParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function updateIdentityDisplayName({
  displayName,
  env,
  identifier,
}: {
  displayName: string;
  env: Env;
  identifier: string;
}): Promise<Response<void>> {
  try {
    await axios({
      baseURL: env.api.url,
      data: { displayName },
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "PATCH",
      url: `${API_VERSION}/identities/${identifier}`,
    });

    return buildSuccessResponse(undefined);
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export const supportedNetworkParser = getStrictParser<SupportedNetwork>()(
  z.object({
    blockchain: z.string(),
    networks: z.tuple([z.string()]).rest(z.string()),
  })
);

export async function getSupportedNetwork({
  env,
  signal,
}: {
  env: Env;
  signal: AbortSignal;
}): Promise<Response<List<SupportedNetwork>>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "GET",
      signal,
      url: `${API_VERSION}/supported-networks`,
    });

    return buildSuccessResponse(getListParser(supportedNetworkParser).parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}
