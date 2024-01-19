import { z } from "zod";

import { List, ResourceMeta } from "src/utils/types";

export function getListParser<Input, Output = Input>(
  parser: z.ZodSchema<Output, z.ZodTypeDef, Input>
) {
  return getStrictParser<unknown[], List<Output>>()(
    z.array(z.unknown()).transform((unknowns) =>
      unknowns.reduce(
        (acc: List<Output>, curr: unknown, index): List<Output> => {
          const parsed = parser.safeParse(curr);

          return parsed.success
            ? {
                ...acc,
                successful: [...acc.successful, parsed.data],
              }
            : {
                ...acc,
                failed: [
                  ...acc.failed,
                  new z.ZodError<Output>(
                    parsed.error.issues.map((issue) => ({
                      ...issue,
                      path: [index, ...issue.path],
                    }))
                  ),
                ],
              };
        },
        { failed: [], successful: [] }
      )
    )
  );
}

const metaParser = getStrictParser<ResourceMeta>()(
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
      meta: metaParser,
    })
  );
}

export const datetimeParser = getStrictParser<string, Date>()(
  z
    .string()
    .datetime()
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
