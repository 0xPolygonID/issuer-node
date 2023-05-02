import z from "zod";

export const processError = (error: unknown) =>
  error instanceof z.ZodError
    ? error
    : error instanceof Error
    ? error.message
    : typeof error === "string"
    ? error
    : "Unknown error";

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

export const credentialSubjectValueErrorToString = (error: string | z.ZodError) =>
  error instanceof z.ZodError
    ? [
        "An error occurred while parsing the value of the credentialSubject:",
        ...processZodError(error).map((e) => `"${e}"`),
      ].join("\n")
    : `An error occurred while processing the value of the credentialSubject:\n"${error}"`;
