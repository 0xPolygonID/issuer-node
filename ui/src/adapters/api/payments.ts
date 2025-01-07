import axios from "axios";
import { z } from "zod";

import { Response, buildErrorResponse, buildSuccessResponse } from "src/adapters";
import { ID, IDParser, Message, buildAuthorizationHeader, messageParser } from "src/adapters/api";
import { datetimeParser, getResourceParser, getStrictParser } from "src/adapters/parsers";
import { Env, PaymentConfigurations, PaymentOption } from "src/domain";
import { API_VERSION } from "src/utils/constants";
import { Resource } from "src/utils/types";

type PaymentOptionInput = Omit<PaymentOption, "modifiedAt" | "createdAt"> & {
  createdAt: string;
  modifiedAt: string;
};

export const paymentOptionParser = getStrictParser<PaymentOptionInput, PaymentOption>()(
  z.object({
    config: z.array(
      z.object({
        amount: z.string(),
        paymentOptionID: z.number(),
        recipient: z.string(),
        signingKeyID: z.string(),
      })
    ),
    createdAt: datetimeParser,
    description: z.string(),
    id: z.string(),
    issuerDID: z.string(),
    modifiedAt: datetimeParser,
    name: z.string(),
  })
);

export const paymentConfigurationsParser = getStrictParser<PaymentConfigurations>()(
  z.record(
    z.object({
      ChainID: z.number(),
      PaymentOption: z.object({
        ContractAddress: z.string(),
        Name: z.string(),
        Type: z.string(),
      }),
      PaymentRails: z.string(),
    })
  )
);

export async function getPaymentOptions({
  env,
  identifier,
  params: { maxResults, page },
  signal,
}: {
  env: Env;
  identifier: string;
  params: {
    maxResults?: number;
    page?: number;
  };
  signal?: AbortSignal;
}): Promise<Response<Resource<PaymentOption>>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "GET",
      params: new URLSearchParams({
        ...(maxResults !== undefined ? { max_results: maxResults.toString() } : {}),
        ...(page !== undefined ? { page: page.toString() } : {}),
      }),
      signal,
      url: `${API_VERSION}/identities/${identifier}/payment/options`,
    });
    return buildSuccessResponse(getResourceParser(paymentOptionParser).parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export type UpsertPaymentOption = Pick<PaymentOption, "name" | "description" | "config">;

export async function createPaymentOption({
  env,
  identifier,
  payload,
}: {
  env: Env;
  identifier: string;
  payload: UpsertPaymentOption;
}): Promise<Response<ID>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      data: payload,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      url: `${API_VERSION}/identities/${identifier}/payment/options`,
    });
    return buildSuccessResponse(IDParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function updatePaymentOption({
  env,
  identifier,
  payload,
  paymentOptionID,
}: {
  env: Env;
  identifier: string;
  payload: UpsertPaymentOption;
  paymentOptionID: string;
}) {
  try {
    await axios({
      baseURL: env.api.url,
      data: payload,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "PATCH",
      url: `${API_VERSION}/identities/${identifier}/payment/options/${paymentOptionID}`,
    });

    return buildSuccessResponse(undefined);
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function getPaymentOption({
  env,
  identifier,
  paymentOptionID,
  signal,
}: {
  env: Env;
  identifier: string;
  paymentOptionID: string;
  signal?: AbortSignal;
}): Promise<Response<PaymentOption>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "GET",
      signal,
      url: `${API_VERSION}/identities/${identifier}/payment/options/${paymentOptionID}`,
    });
    return buildSuccessResponse(paymentOptionParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function deletePaymentOption({
  env,
  identifier,
  paymentOptionID,
}: {
  env: Env;
  identifier: string;
  paymentOptionID: string;
}): Promise<Response<Message>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "DELETE",
      url: `${API_VERSION}/identities/${identifier}/payment/options/${paymentOptionID}`,
    });
    return buildSuccessResponse(messageParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function getPaymentConfigurations({
  env,
  signal,
}: {
  env: Env;
  signal?: AbortSignal;
}): Promise<Response<PaymentConfigurations>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "GET",
      signal,
      url: `${API_VERSION}/payment/settings`,
    });
    return buildSuccessResponse(paymentConfigurationsParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}
