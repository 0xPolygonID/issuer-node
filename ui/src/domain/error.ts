import { AxiosError, Cancel } from "axios";
import { ZodError } from "zod";

type CancelError = {
  error: Cancel;
  message: string;
  type: "cancel-error";
};

type CustomError = {
  message: string;
  type: "custom-error";
};

type GeneralError = {
  error: Error;
  message: string;
  type: "general-error";
};

type ParseError = {
  error: ZodError;
  message: string;
  type: "parse-error";
};

type RequestError = {
  error: AxiosError;
  message: string;
  type: "request-error";
};

type UnknownError = {
  error: unknown;
  message: string;
  type: "unknown-error";
};

export type AppError =
  | CancelError
  | CustomError
  | GeneralError
  | ParseError
  | RequestError
  | UnknownError;
