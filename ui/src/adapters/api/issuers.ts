import axios from "axios";
import { z } from "zod";

import { Response, buildErrorResponse, buildSuccessResponse } from "src/adapters";
import { buildAuthorizationHeader } from "src/adapters/api";
import { getListParser, getStrictParser } from "src/adapters/parsers";
import {
  CredentialStatusType,
  Env,
  Issuer,
  IssuerIdentifier,
  IssuerInfo,
  IssuerType,
  Method,
  SupportedNetwork,
} from "src/domain";

import { API_VERSION } from "src/utils/constants";
import { List } from "src/utils/types";

const apiIssuerParser = getStrictParser<Issuer>()(
  z.object({
    blockchain: z.string(),
    credentialStatusType: z.nativeEnum(CredentialStatusType),
    displayName: z.string(),
    identifier: z.string(),
    method: z.nativeEnum(Method),
    network: z.string(),
  })
);

export const identifierParser = getStrictParser<IssuerIdentifier>()(z.string());

export async function getIssuers({
  env,
  signal,
}: {
  env: Env;
  signal: AbortSignal;
}): Promise<Response<List<Issuer>>> {
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

    return buildSuccessResponse(getListParser(apiIssuerParser).parse(response.data || []));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export type CreateIssuer = {
  blockchain: string;
  credentialStatusType: CredentialStatusType;
  displayName: string;
  method: string;
  network: string;
  type: IssuerType;
};

type CreatedIssuer = {
  identifier: string;
};

export const createIssuerParser = getStrictParser<CreatedIssuer>()(
  z.object({
    identifier: z.string(),
  })
);

export async function createIssuer({
  env,
  payload,
}: {
  env: Env;
  payload: CreateIssuer;
}): Promise<Response<CreatedIssuer>> {
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

    return buildSuccessResponse(createIssuerParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export const issuerDetailsParser = getStrictParser<IssuerInfo>()(
  z.object({
    credentialStatusType: z.nativeEnum(CredentialStatusType),
    displayName: z.string(),
    identifier: z.string(),
    keyType: z.nativeEnum(IssuerType),
  })
);

export async function getIssuerDetails({
  env,
  identifier,
  signal,
}: {
  env: Env;
  identifier: IssuerIdentifier;
  signal?: AbortSignal;
}): Promise<Response<IssuerInfo>> {
  try {
    const response = await axios({
      baseURL: env.api.url,

      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "GET",
      signal,
      url: `${API_VERSION}/identities/${identifier}/details`,
    });

    return buildSuccessResponse(issuerDetailsParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function updateIssuerDisplayName({
  displayName,
  env,
  identifier,
}: {
  displayName: string;
  env: Env;
  identifier: IssuerIdentifier;
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
    networks: z.array(z.string()).nonempty(),
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
