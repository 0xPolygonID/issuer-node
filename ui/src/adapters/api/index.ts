import axios from "axios";
import z from "zod";

import { getStrictParser } from "src/adapters/parsers";
import { Env } from "src/domain";
import { processZodError } from "src/utils/error";

export type GetAll<T> = {
  failed: z.ZodError<T>[];
  successful: T[];
};

export interface APIError {
  message: string;
  status?: number;
}

interface APIErrorResponse {
  error: APIError;
  isSuccessful: false;
}

interface APISuccessfulResponse<D> {
  data: D;
  isSuccessful: true;
}

export type APIResponse<D> = APISuccessfulResponse<D> | APIErrorResponse;

export interface ResultAccepted<D> {
  data: D;
  status: 202;
}

export interface ResultOK<D> {
  data: D;
  status: 200;
}

export interface ResultCreated<D> {
  data: D;
  status: 201;
}

interface ResponseError {
  data: { message: string };
  status: number;
}

const responseErrorParser = getStrictParser<ResponseError>()(
  z.object({
    data: z.object({ message: z.string() }),
    status: z.number(),
  })
);

export interface ID {
  id: string;
}

export const IDParser = getStrictParser<ID>()(z.object({ id: z.string() }));

export function buildAPIError(error: unknown): APIError {
  if (axios.isCancel(error)) {
    return { message: error.toString(), status: 0 };
  }

  if (axios.isAxiosError(error)) {
    try {
      // This is a UI API error.
      const { data, status } = responseErrorParser.parse(error.response);
      const { message } = data;

      return { message, status };
    } catch (e) {
      // This catches a CORS or other network error.
      return { message: error.message };
    }
  }

  if (error instanceof z.ZodError) {
    return { message: processZodError(error).join("\n") };
  }

  if (error instanceof Error) {
    // This is an application-level error.
    return { message: error.toString() };
  }

  // This shouldn't happen (catch-all).
  console.error(error);
  return { message: "Unknown error" };
}

export function buildAuthorizationHeader(env: Env) {
  return `Basic ${window.btoa(`${env.api.username}:${env.api.password}`)}`;
}

export const resultOKMessage = getStrictParser<ResultOK<{ message: string }>>()(
  z.object({
    data: z.object({
      message: z.string(),
    }),
    status: z.literal(200),
  })
);
