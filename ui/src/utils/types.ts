import { ZodSchema, ZodTypeDef } from "zod";

export type AsyncTask<Data, Error> =
  | AsyncTaskFailed<Error>
  | AsyncTaskLoading
  | AsyncTaskPending
  | AsyncTaskReloading<Data>
  | AsyncTaskSuccessful<Data>;

interface AsyncTaskFailed<Error> {
  error: Error;
  status: "failed";
}

interface AsyncTaskLoading {
  status: "loading";
}

interface AsyncTaskPending {
  status: "pending";
}

interface AsyncTaskReloading<Data> {
  data: Data;
  status: "reloading";
}

interface AsyncTaskSuccessful<Data> {
  data: Data;
  status: "successful";
}

export function isAsyncTaskDataAvailable<Data, Error>(
  task: AsyncTask<Data, Error>
): task is AsyncTaskReloading<Data> | AsyncTaskSuccessful<Data> {
  return task.status === "reloading" || task.status === "successful";
}

export function isAsyncTaskStarting<Data, Error>(
  task: AsyncTask<Data, Error>
): task is AsyncTaskLoading | AsyncTaskPending {
  return task.status === "loading" || task.status === "pending";
}

export function hasAsyncTaskFailed<Data, Error>(
  task: AsyncTask<Data, Error>
): task is AsyncTaskFailed<Error> {
  return task.status === "failed";
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
    ZodSchema<Output, ZodTypeDef, Input>,
    ZodSchema<ParserOutput, ZodTypeDef, ParserInput>
  > extends true
    ? // 2. then if the output types match exactly,
      Exact<Required<Output>, Required<ParserOutput>> extends true
      ? // 3. then if the input types match exactly,
        Exact<Required<Input>, Required<ParserInput>> extends true
        ? // 3. return type as a ZodSchema parser (it is certain that I/O and parser I/O are identical).
          ZodSchema<ParserOutput, ZodTypeDef, ParserInput>
        : // 3. fail
          never
      : // 2. fail
        never
    : // 1. fail
      never
) => ZodSchema<Output, ZodTypeDef, Input> {
  return (parser: ZodSchema) => parser;
}
