import axios from "axios";
import dayjs from "dayjs";
import { z } from "zod";

import { RequestResponse } from "src/adapters";
import { buildAuthorizationHeader } from "src/adapters/api";
import { getListParser, getStrictParser } from "src/adapters/parsers";
import { Env, IssuerStatus, Transaction, TransactionStatus } from "src/domain";
import { API_VERSION } from "src/utils/constants";
import { buildAppError } from "src/utils/error";
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

export async function publishState({ env }: { env: Env }): Promise<RequestResponse<boolean>> {
  try {
    await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      url: `${API_VERSION}/state/publish`,
    });

    return { data: true, success: true };
  } catch (error) {
    return { error: buildAppError(error), success: false };
  }
}

export async function retryPublishState({ env }: { env: Env }): Promise<RequestResponse<boolean>> {
  try {
    await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      url: `${API_VERSION}/state/retry`,
    });

    return { data: true, success: true };
  } catch (error) {
    return { error: buildAppError(error), success: false };
  }
}

export async function getStatus({
  env,
  signal,
}: {
  env: Env;
  signal?: AbortSignal;
}): Promise<RequestResponse<IssuerStatus>> {
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
    const data = issuerStatusParser.parse(response.data);

    return { data, success: true };
  } catch (error) {
    return { error: buildAppError(error), success: false };
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
}): Promise<RequestResponse<List<Transaction>>> {
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
    const data = getListParser(transactionParser).parse(response.data);

    return {
      data: {
        failed: data.failed,
        successful: data.successful.sort(
          ({ publishDate: a }, { publishDate: b }) => dayjs(b).unix() - dayjs(a).unix()
        ),
      },
      success: true,
    };
  } catch (error) {
    return { error: buildAppError(error), success: false };
  }
}
