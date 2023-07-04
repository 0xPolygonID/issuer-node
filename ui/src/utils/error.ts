import { message } from "antd";
import { isAxiosError, isCancel } from "axios";
import z from "zod";
import { getStrictParser } from "src/adapters/parsers";
import { AppError } from "src/domain";

function processZodError<T>(error: z.ZodError<T>, init: string[] = []) {
  return error.errors.reduce((mainAcc, issue): string[] => {
    switch (issue.code) {
      case "invalid_union": {
        return [
          ...mainAcc,
          ...issue.unionErrors.reduce(
            (innerAcc: string[], current: z.ZodError<T>): string[] => [
              ...innerAcc,
              ...processZodError(current),
            ],
            []
          ),
        ];
      }

      default: {
        const errorMsg = issue.path.length
          ? `${issue.message} at ${issue.path.join(".")}`
          : issue.message;
        return [...mainAcc, errorMsg];
      }
    }
  }, init);
}

export function notifyError(error: AppError, compact = false): void {
  if (!compact && error.type === "parse-error") {
    notifyParseError(error.error);
  } else {
    void message.error(error.message);
  }
}

export function notifyParseError(error: z.ZodError): void {
  processZodError(error).forEach((error) => void message.error(error));
}

export function notifyParseErrors(errors: z.ZodError[]): void {
  errors.forEach(notifyParseError);
}

const messageParser = getStrictParser<{ message: string }>()(z.object({ message: z.string() }));

export function buildAppError(error: unknown): AppError {
  if (typeof error === "string") {
    return {
      message: error,
      type: "custom-error",
    };
  } else if (isCancel(error)) {
    return {
      error,
      message: error.message
        ? `The request has been aborted. ${error.message}`
        : "The request has been aborted.",
      type: "cancel-error",
    };
  } else if (isAxiosError(error)) {
    const parsedMessage = messageParser.safeParse(error.response?.data);

    return {
      error,
      message: parsedMessage.success
        ? `${error.message}: ${parsedMessage.data.message}`
        : error.message,
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

export const envErrorToString = (error: AppError) =>
  [
    "An error occurred while reading the environment variables:",
    error.message,
    "Please provide valid environment variables.",
  ].join("\n");

export const credentialSubjectValueErrorToString = (error: AppError) =>
  [
    error.type === "parse-error" || error.type === "custom-error"
      ? "An error occurred while parsing the value of the credentialSubject:"
      : "An error occurred while processing the value of the credentialSubject",
    error.message,
    "Please try again.",
  ].join("\n");

export const jsonSchemaErrorToString = (error: AppError) =>
  [
    error.type === "parse-error" || error.type === "custom-error"
      ? "An error occurred while parsing the JSON Schema:"
      : "An error occurred while downloading the JSON Schema:",
    error.message,
    "Please try again.",
  ].join("\n");

export const jsonLdContextErrorToString = (error: AppError) =>
  [
    error.type === "parse-error" || error.type === "custom-error"
      ? "An error occurred while parsing the JSON LD Type referenced in this schema:"
      : "An error occurred while downloading the JSON LD Type referenced in this schema:",
    error.message,
    "Please try again.",
  ].join("\n");
