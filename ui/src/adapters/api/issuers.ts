import axios from "axios";
import { z } from "zod";

import { Response, buildErrorResponse, buildSuccessResponse } from "src/adapters";
import { buildAuthorizationHeader } from "src/adapters/api";
import { datetimeParser, getListParser, getStrictParser } from "src/adapters/parsers";
import { Env } from "src/domain";

import {
  AuthBJJCredentialStatus,
  Identifier,
  Issuer,
  IssuerFormData,
  IssuerState,
} from "src/domain/identifier";
import { API_VERSION } from "src/utils/constants";
import { List } from "src/utils/types";

type ApiIssuerStateInput = Omit<IssuerState, "state"> & {
  state: Omit<IssuerState["state"], "createdAt" | "modifiedAt"> & {
    createdAt: string;
    modifiedAt: string;
  };
};

const apiIssuerParser = getStrictParser<Issuer>()(
  z.object({
    authBJJCredentialStatus: z.enum([
      AuthBJJCredentialStatus.Iden3OnchainSparseMerkleTreeProof2023,
      AuthBJJCredentialStatus.Iden3ReverseSparseMerkleTreeProof,
      AuthBJJCredentialStatus["Iden3commRevocationStatusV1.0"],
    ]),
    blockchain: z.string(),
    identifier: z.string(),
    method: z.string(),
    network: z.string(),
  })
);

export const identifierParser = getStrictParser<Identifier>()(z.string().nullable());

const apiIssuerStateParser = getStrictParser<ApiIssuerStateInput, IssuerState>()(
  z.object({
    address: z.string(),
    identifier: z.string(),
    state: z.object({
      blockNumber: z.number().optional(),
      blockTimestamp: z.number().optional(),
      claimsTreeRoot: z.string(),
      createdAt: datetimeParser,
      identifier: z.string().optional(),
      modifiedAt: datetimeParser,
      previousState: z.string().optional(),
      revocationTreeRoot: z.string().optional(),
      rootOfRoots: z.string().optional(),
      state: z.string(),
      stateID: z.number().optional(),
      status: z.string(),
      txID: z.string().optional(),
    }),
  })
);

export async function getApiIssuers({
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

export async function createIssuer({
  env,
  payload,
}: {
  env: Env;
  payload: IssuerFormData;
}): Promise<Response<IssuerState>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      data: { didMetadata: payload },
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      url: `${API_VERSION}/identities`,
    });

    return buildSuccessResponse(apiIssuerStateParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}
