import axios from "axios";
import { z } from "zod";

import { Response, buildErrorResponse, buildSuccessResponse } from "src/adapters";
import { ID, IDParser, Message, buildAuthorizationHeader, messageParser } from "src/adapters/api";
import { datetimeParser, getListParser, getStrictParser } from "src/adapters/parsers";
import { ApiSchema, Env, Json, Request } from "src/domain";
import { DataSchema } from "src/domain/dataSchema";
import { API_VERSION, QUERY_SEARCH_PARAM, STATUS_SEARCH_PARAM } from "src/utils/constants";
import { List } from "src/utils/types";

type RequestInput = Omit<Request, "proofTypes" | "createdAt" | "expiresAt"> & {
  createdAt: string;
  expiresAt: string | null;
};
type ApiSchemaInput = Omit<ApiSchema, "createdAt"> & {
  createdAt: string;
};
const User = localStorage.getItem("user");
const apiSchemaParser = getStrictParser<ApiSchemaInput, ApiSchema>()(
  z.object({
    bigInt: z.string(),
    createdAt: datetimeParser,
    description: z.string().nullable(),
    hash: z.string(),
    id: z.string(),
    title: z.string().nullable(),
    type: z.string(),
    url: z.string(),
    version: z.string().nullable(),
  })
);

export const RequestParser = getStrictParser<RequestInput, Request>()(
  z.object({
    Active: z.boolean(),
    age: z.string(),
    created_at: datetimeParser,
    credential_type: z.string(),
    id: z.string(),
    IssuerId: z.string(),
    modified_at: datetimeParser.nullable(),
    proof_id: z.string(),
    proof_type: z.string(),
    request_status: z.string(),
    request_type: z.string(),
    role_type: z.string(),
    schemaID: z.string(),
    source: z.string(),
    userDID: z.string(),
    verifier_status: z.string(),
    wallet_status: z.string(),
  })
);

export type RequestStatus = "all" | "revoked" | "expired";

export const requestStatusParser = getStrictParser<RequestStatus>()(
  z.union([z.literal("all"), z.literal("revoked"), z.literal("expired")])
);

export async function getRequest({
  env,
  signal,
}: {
  RequestID: string;
  env: Env;
  signal?: AbortSignal;
}): Promise<Response<Request>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "GET",
      signal,
      url: `${API_VERSION}/requests`,
    });
    return buildSuccessResponse(RequestParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function getRequests({
  env,
  params: { did, query, status },
  signal,
}: {
  env: Env;
  params: {
    did?: string;
    query?: string;
    status?: RequestStatus;
  };
  signal?: AbortSignal;
}): Promise<Response<List<Request>>> {
  try {
    let response;
    let response1;
    // let response2;
    if (User === "verifier" || User === "issuer") {
      response = await axios({
        baseURL: env.api.url,
        data:
          query !== undefined
            ? User === "issuer"
              ? { Request_type: "GenerateNewVC", UserDID: query }
              : { Request_type: "VerifyVC", UserDID: query }
            : User === "issuer"
            ? { Request_type: "GenerateNewVC" }
            : { Request_type: "VerifyVC" },
        headers: {
          Authorization: buildAuthorizationHeader(env),
        },
        method: "POST",
        params: new URLSearchParams({
          ...(did !== undefined ? { did } : {}),
          ...(query !== undefined ? { [QUERY_SEARCH_PARAM]: query } : {}),
          ...(status !== undefined && status !== "all" ? { [STATUS_SEARCH_PARAM]: status } : {}),
        }),
        signal,
        url: query !== undefined ? `${API_VERSION}/requests/all` : `${API_VERSION}/requests/bytype`,
      });
    } else {
      response1 = await axios({
        baseURL: env.api.url,
        data: { Request_type: "GenerateNewVC", UserDID: query },
        headers: {
          Authorization: buildAuthorizationHeader(env),
        },
        method: "POST",
        params: new URLSearchParams({
          ...(did !== undefined ? { did } : {}),
          ...(query !== undefined ? { [QUERY_SEARCH_PARAM]: query } : {}),
          ...(status !== undefined && status !== "all" ? { [STATUS_SEARCH_PARAM]: status } : {}),
        }),
        signal,
        url: `${API_VERSION}/requests/all`,
      });
      // response2 = await axios({
      //   baseURL: env.api.url,
      //   data: { Request_type: "VerifyVC", UserDID: query },
      //   headers: {
      //     Authorization: buildAuthorizationHeader(env),
      //   },
      //   method: "POST",
      //   params: new URLSearchParams({
      //     ...(did !== undefined ? { did } : {}),
      //     ...(query !== undefined ? { [QUERY_SEARCH_PARAM]: query } : {}),
      //     ...(status !== undefined && status !== "all" ? { [STATUS_SEARCH_PARAM]: status } : {}),
      //   }),
      //   signal,
      //   url: `${API_VERSION}/requests/all`,
      // });
      response = response1;

      // response.data = [...response.data, ...response2.data];
    }

    return buildSuccessResponse(
      getListParser(RequestParser)
        .transform(({ failed, successful }) => ({
          failed,
          successful: successful.sort((a, b) => b.created_at.getTime() - a.created_at.getTime()),
        }))
        .parse(response.data)
    );
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export type CreateRequest = {
  RequestSchema: string;
  RequestSubject: Json;
  expiration: string | null;
  mtProof: boolean;
  signatureProof: boolean;
  type: string;
};

export async function createRequest({
  env,
  payload,
}: {
  env: Env;
  payload: CreateRequest;
}): Promise<Response<ID>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      data: payload,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      url: `${API_VERSION}/Requests`,
    });
    return buildSuccessResponse(IDParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function revokeRequest({
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
      url: `${API_VERSION}/Requests/revoke/${nonce}`,
    });
    return buildSuccessResponse(messageParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function deleteRequest({
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
      url: `${API_VERSION}/Requests/${id}`,
    });
    return buildSuccessResponse(messageParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function issueCredentialRequest({
  dataSchema,
  env,
}: {
  dataSchema: DataSchema;
  env: Env;
}): Promise<Response<ID>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      data: dataSchema,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      url: `${API_VERSION}/generateVC`,
    });
    return buildSuccessResponse(IDParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function getSchema({
  env,
  schemaID,
}: {
  env: Env;
  schemaID: string;
}): Promise<Response<ApiSchema>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "GET",
      url: `${API_VERSION}/schemas/${schemaID}`,
    });
    return buildSuccessResponse(apiSchemaParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}
