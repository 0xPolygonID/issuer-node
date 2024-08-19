import axios from "axios";
import { z } from "zod";

import { Response, buildErrorResponse, buildSuccessResponse } from "src/adapters";
import { buildAuthorizationHeader } from "src/adapters/api";
import { getListParser, getStrictParser } from "src/adapters/parsers";
import { AuthBJJCredentialStatus, Env, Issuer, IssuerIdentifier, IssuerType } from "src/domain";

import { API_VERSION } from "src/utils/constants";
import { List } from "src/utils/types";

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
  method: string;
  network: string;
} & (
  | {
      authBJJCredentialStatus: AuthBJJCredentialStatus;
      type: IssuerType.BJJ;
    }
  | {
      type: IssuerType.ETH;
    }
);

export async function createIssuer({
  env,
  payload,
}: {
  env: Env;
  payload: CreateIssuer;
}): Promise<Response<void>> {
  try {
    await axios({
      baseURL: env.api.url,
      data: { didMetadata: payload },
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      url: `${API_VERSION}/identities`,
    });

    return buildSuccessResponse(undefined);
  } catch (error) {
    return buildErrorResponse(error);
  }
}
