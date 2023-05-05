import axios from "axios";
import { z } from "zod";

import { Response, buildErrorResponse, buildSuccessResponse } from "src/adapters";
import { buildAuthorizationHeader } from "src/adapters/api";
import { getListParser, getStrictParser } from "src/adapters/parsers";
import { Env, IssuerStatus, Transaction, TransactionStatus } from "src/domain";
import { API_VERSION } from "src/utils/constants";
import { List } from "src/utils/types";

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
    publishDate: z.coerce.date(z.string().datetime()),
    state: z.string(),
    status: transactionStatusParser,
    txID: z.string(),
  })
);

export async function publishState({ env }: { env: Env }): Promise<Response<null>> {
  try {
    await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      url: `${API_VERSION}/state/publish`,
    });
    return buildSuccessResponse(null);
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function retryPublishState({ env }: { env: Env }): Promise<Response<null>> {
  try {
    await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      url: `${API_VERSION}/state/retry`,
    });
    return buildSuccessResponse(null);
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function getStatus({
  env,
  signal,
}: {
  env: Env;
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
      url: `${API_VERSION}/state/status`,
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
  signal,
}: {
  env: Env;
  signal?: AbortSignal;
}): Promise<Response<List<Transaction>>> {
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
    return buildSuccessResponse(
      getListParser(transactionParser)
        .transform(({ failed, successful }) => ({
          failed,
          successful: successful.sort((a, b) => b.publishDate.getTime() - a.publishDate.getTime()),
        }))
        .parse(response.data)
    );
  } catch (error) {
    return buildErrorResponse(error);
  }
}
