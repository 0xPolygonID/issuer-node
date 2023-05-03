import axios from "axios";
import z from "zod";

import { getStrictParser } from "src/adapters/parsers";
import { AppError, Env } from "src/domain";
import { processZodError } from "src/utils/error";

interface RequestErrorResponse {
  error: AppError;
  isSuccessful: false;
}

interface RequestSuccessfulResponse<D> {
  data: D;
  isSuccessful: true;
}

export type RequestResponse<D> = RequestSuccessfulResponse<D> | RequestErrorResponse;

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

export interface ID {
  id: string;
}

export const IDParser = getStrictParser<ID>()(z.object({ id: z.string() }));

export function buildAppError(error: unknown): AppError {
  if (axios.isCancel(error)) {
    return {
      error,
      message: error.message
        ? `The request has been aborted. ${error.message}`
        : "The request has been aborted.",
      type: "cancel-error",
    };
  } else if (axios.isAxiosError(error)) {
    return {
      error,
      message: error.message,
      type: "request-error",
    };
  } else if (error instanceof z.ZodError) {
    return {
      error,
      message: processZodError(error).join("\n"),
      type: "parse-error",
    };
  } else if (error instanceof Error) {
    return {
      error,
      message: error.message,
      type: "general-error",
    };
  } else {
    return {
      error,
      message: "Unknown error",
      type: "unknown-error",
    };
  }
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
