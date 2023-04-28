import z from "zod";

import { List } from "src/utils/types";

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

// The following was implemented due to a perceived limitation of Zod:
// https://github.com/colinhacks/zod/issues/652

// Asserts whether 2 types are exactly the same.
type Exact<T, U> = [T, U] extends [U, T] ? true : false;

// A function that accepts two types then generates a function that accepts a parser for them if the types match,
// otherwise it produces a static error.
export function getStrictParser<Input, Output = Input>(): <
  ParserOutput,
  ParserInput = ParserOutput
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
