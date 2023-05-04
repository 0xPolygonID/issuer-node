import { AxiosError, Cancel } from "axios";
import { ZodError } from "zod";

interface CancelError {
  error: Cancel;
  message: string;
  type: "cancel-error";
}

interface CustomError {
  message: string;
  type: "custom-error";
}

interface GeneralError {
  error: Error;
  message: string;
  type: "general-error";
}

interface ParseError {
  error: ZodError;
  message: string;
  type: "parse-error";
}

interface RequestError {
  error: AxiosError;
  message: string;
  type: "request-error";
}

interface UnknownError {
  error: unknown;
  message: string;
  type: "unknown-error";
}

export type AppError =
  | CancelError
  | CustomError
  | GeneralError
  | ParseError
  | RequestError
  | UnknownError;
