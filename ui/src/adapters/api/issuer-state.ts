import axios from "axios";
import { z } from "zod";

import {
  APIResponse,
  HTTPStatusSuccess,
  ResultOK,
  buildAPIError,
  buildAuthorizationHeader,
} from "src/adapters/api";
import { getStrictParser } from "src/adapters/parsers";
import { Env, IssuerStatus, Transaction, TransactionStatus } from "src/domain";
import { API_VERSION } from "src/utils/constants";

const transactionStatusParser = getStrictParser<TransactionStatus>()(
  z.union([
    z.literal("created"),
    z.literal("failed"),
    z.literal("pending"),
    z.literal("published"),
    z.literal("transacted"),
  ])
);

const transactionParser = getStrictParser<Transaction>()(
  z.object({
    id: z.number(),
    publishDate: z.coerce.date(),
    state: z.string(),
    status: transactionStatusParser,
    txID: z.string(),
  })
);

export async function publishState({ env }: { env: Env }): Promise<APIResponse<boolean>> {
  try {
    await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      url: `${API_VERSION}/state/publish`,
    });

    return { data: true, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

export async function retryPublishState({ env }: { env: Env }): Promise<APIResponse<boolean>> {
  try {
    await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      url: `${API_VERSION}/state/retry`,
    });

    return { data: true, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

export async function getStatus({
  env,
  signal,
}: {
  env: Env;
  signal?: AbortSignal;
}): Promise<APIResponse<IssuerStatus>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "GET",
      signal,
      url: `${API_VERSION}/state/status`,
    });
    const { data } = resultOKStatusParser.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

const resultOKStatusParser = getStrictParser<ResultOK<IssuerStatus>>()(
  z.object({
    data: z.object({ pendingActions: z.boolean() }),
    status: z.literal(HTTPStatusSuccess.OK),
  })
);

export async function getTransactions({
  env,
  signal,
}: {
  env: Env;
  signal?: AbortSignal;
}): Promise<APIResponse<Transaction[]>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "GET",
      signal,
      url: `${API_VERSION}/state/transactions`,
    });
    const { data } = resultOKTransactionsParser.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

const resultOKTransactionsParser = getStrictParser<ResultOK<Transaction[]>>()(
  z.object({
    data: z.array(transactionParser),
    status: z.literal(HTTPStatusSuccess.OK),
  })
);
