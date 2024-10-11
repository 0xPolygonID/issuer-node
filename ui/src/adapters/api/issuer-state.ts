import axios from "axios";
import { z } from "zod";

import { Response, buildErrorResponse, buildSuccessResponse } from "src/adapters";
import { Sorter, buildAuthorizationHeader, serializeSorters } from "src/adapters/api";
import { datetimeParser, getResourceParser, getStrictParser } from "src/adapters/parsers";
import { Env, IssuerStatus, Transaction, TransactionStatus } from "src/domain";
import { API_VERSION, QUERY_SEARCH_PARAM } from "src/utils/constants";
import { Resource } from "src/utils/types";

const transactionStatusParser = getStrictParser<TransactionStatus>()(
  z.union([
    z.literal("created"),
    z.literal("failed"),
    z.literal("pending"),
    z.literal("published"),
    z.literal("transacted"),
  ])
);

type TransactionInput = Omit<Transaction, "publishDate"> & {
  publishDate: string;
};

const transactionParser = getStrictParser<TransactionInput, Transaction>()(
  z.object({
    id: z.number(),
    publishDate: datetimeParser,
    state: z.string(),
    status: transactionStatusParser,
    txID: z.string(),
  })
);

export async function publishState({
  env,
  identifier,
}: {
  env: Env;
  identifier: string;
}): Promise<Response<null>> {
  try {
    await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      url: `${API_VERSION}/identities/${identifier}/state/publish`,
    });
    return buildSuccessResponse(null);
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function retryPublishState({
  env,
  identifier,
}: {
  env: Env;
  identifier: string;
}): Promise<Response<null>> {
  try {
    await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      url: `${API_VERSION}/identities/${identifier}/state/retry`,
    });
    return buildSuccessResponse(null);
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function getStatus({
  env,
  identifier,
  signal,
}: {
  env: Env;
  identifier: string;
  signal?: AbortSignal;
}): Promise<Response<IssuerStatus>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "GET",
      signal,
      url: `${API_VERSION}/identities/${identifier}/state/status`,
    });
    return buildSuccessResponse(issuerStatusParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

const issuerStatusParser = getStrictParser<IssuerStatus>()(
  z.object({ pendingActions: z.boolean() })
);

export async function getTransactions({
  env,
  identifier,
  params: { maxResults, page, query, sorters },
  signal,
}: {
  env: Env;
  identifier: string;
  params: {
    maxResults?: number;
    page?: number;
    query?: string;
    sorters?: Sorter[];
  };
  signal?: AbortSignal;
}): Promise<Response<Resource<Transaction>>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "GET",
      params: new URLSearchParams({
        ...(query !== undefined ? { [QUERY_SEARCH_PARAM]: query } : {}),
        ...(maxResults !== undefined ? { max_results: maxResults.toString() } : {}),
        ...(page !== undefined ? { page: page.toString() } : {}),
        ...(sorters !== undefined && sorters.length ? { sort: serializeSorters(sorters) } : {}),
      }),
      signal,
      url: `${API_VERSION}/identities/${identifier}/state/transactions`,
    });

    return buildSuccessResponse(getResourceParser(transactionParser).parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}
