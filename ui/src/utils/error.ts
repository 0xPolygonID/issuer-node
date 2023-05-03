import { message } from "antd";
import z from "zod";
import { AppError } from "src/domain";

export function notifyParseError(error: z.ZodError): void {
  processZodError(error).forEach((error) => void message.error(error));
}

export function notifyParseErrors(errors: z.ZodError[]): void {
  errors.forEach(notifyParseError);
}

export function processZodError<T>(error: z.ZodError<T>, init: string[] = []) {
  return error.errors.reduce((mainAcc, issue): string[] => {
    switch (issue.code) {
      case "invalid_union": {
        return [
          ...mainAcc,
          ...issue.unionErrors.reduce(
            (innerAcc: string[], current: z.ZodError<T>): string[] => [
              ...innerAcc,
              ...processZodError(current, mainAcc),
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
      ? "An error occurred while parsing the json schema:"
      : "An error occurred while downloading the json schema:",
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
