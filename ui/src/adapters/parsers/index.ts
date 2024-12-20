import { message } from "antd";
import { MessageType } from "antd/es/message/interface";
import { isAxiosError, isCancel } from "axios";
import { z } from "zod";

import { AppError } from "src/domain";
import { List, ResourceMeta } from "src/utils/types";

export function getListParser<Input, Output = Input>(
  parser: z.ZodSchema<Output, z.ZodTypeDef, Input>
) {
  return getStrictParser<unknown[], List<Output>>()(
    z.array(z.unknown()).transform((unknowns) =>
      unknowns.reduce(
        (acc: List<Output>, curr: unknown, index): List<Output> => {
          const parsed = parser.safeParse(curr);

          if (parsed.success) {
            return {
              ...acc,
              successful: [...acc.successful, parsed.data],
            };
          } else {
            const error = new z.ZodError<Output>(
              parsed.error.issues.map((issue) => ({
                ...issue,
                path: [index, ...issue.path],
              }))
            );
            return {
              ...acc,
              failed: [
                ...acc.failed,
                {
                  error,
                  message: processZodError(error).join("\n"),
                  type: "parse-error",
                },
              ],
            };
          }
        },
        { failed: [], successful: [] }
      )
    )
  );
}

const resourceMetaParser = getStrictParser<ResourceMeta>()(
  z.object({
    max_results: z.number().int().min(1),
    page: z.number().int().min(1),
    total: z.number().int().min(0),
  })
);

export function getResourceParser<Input, Output = Input>(
  parser: z.ZodSchema<Output, z.ZodTypeDef, Input>
) {
  return getStrictParser<
    { items: unknown[]; meta: ResourceMeta },
    { items: List<Output>; meta: ResourceMeta }
  >()(
    z.object({
      items: getListParser(parser),
      meta: resourceMetaParser,
    })
  );
}

export const datetimeParser = getStrictParser<string, Date>()(
  z
    .string()
    .datetime({ offset: true })
    .transform((datetime, context) => {
      const parsedDate = z.coerce.date().safeParse(datetime);

      if (parsedDate.success) {
        return parsedDate.data;
      } else {
        parsedDate.error.issues.map(context.addIssue);
        return z.NEVER;
      }
    })
);

export const positiveIntegerFromStringParser = getStrictParser<string, number>()(
  z.string().transform((value, context) => {
    const trimmed = value.trim();
    const valueToParse = trimmed === "" ? undefined : Number(trimmed);
    const parsedNumber = z.number().int().min(1).safeParse(valueToParse);
    if (parsedNumber.success) {
      return parsedNumber.data;
    } else {
      parsedNumber.error.issues.map(context.addIssue);
      return z.NEVER;
    }
  })
);

// The following was implemented due to a perceived limitation of Zod:
// https://github.com/colinhacks/zod/issues/652

// Asserts whether 2 types are exactly the same.
type Exact<T, U> = [T, U] extends [U, T] ? true : false;

// A function that accepts two types then generates a function that accepts a parser for them if the types match,
// otherwise it produces a static error.
export function getStrictParser<Input, Output = Input>(): <
  ParserOutput,
  ParserInput = ParserOutput,
>(
  // 1. If the parser's input is exactly the same as the output of the provided zod schema,
  parser: Exact<
    z.ZodSchema<Output, z.ZodTypeDef, Input>,
    z.ZodSchema<ParserOutput, z.ZodTypeDef, ParserInput>
  > extends true
    ? // 2. then if the output types match exactly,
      Exact<Required<Output>, Required<ParserOutput>> extends true
      ? // 3. then if the input types match exactly,
        Exact<Required<Input>, Required<ParserInput>> extends true
        ? // 3. return type as a z.ZodSchema parser (it is certain that Input/Output and parser Input/Output are identical).
          z.ZodSchema<ParserOutput, z.ZodTypeDef, ParserInput>
        : // 3. fail
          never
      : // 2. fail
        never
    : // 1. fail
      never
) => z.ZodSchema<Output, z.ZodTypeDef, Input> {
  return (parser: z.ZodSchema) => parser;
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

export function notifyError(error: AppError, compact = false): MessageType[] {
  if (!compact && error.type === "parse-error") {
    return notifyParseError(error.error);
  } else {
    return [message.error(error.message)];
  }
}

export function notifyParseError(error: z.ZodError): MessageType[] {
  return processZodError(error).map((error) => message.error(error));
}

export function notifyErrors(errors: AppError[]): MessageType[] {
  return errors.reduce(
    (acc: MessageType[], curr) => [
      ...acc,
      ...(curr.type === "parse-error" ? notifyParseError(curr.error) : notifyError(curr)),
    ],
    []
  );
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
