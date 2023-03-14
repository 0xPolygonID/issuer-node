import { ZodSchema, ZodTypeDef } from "zod";

export type AsyncTask<D, E> =
  | AsyncTaskFailed<E>
  | AsyncTaskLoading
  | AsyncTaskPending
  | AsyncTaskReloading<D>
  | AsyncTaskSuccessful<D>;

interface AsyncTaskFailed<E> {
  error: E;
  status: "failed";
}

interface AsyncTaskLoading {
  status: "loading";
}

interface AsyncTaskPending {
  status: "pending";
}

interface AsyncTaskReloading<D> {
  data: D;
  status: "reloading";
}

interface AsyncTaskSuccessful<D> {
  data: D;
  status: "successful";
}

export function isAsyncTaskDataAvailable<D, E>(
  task: AsyncTask<D, E>
): task is AsyncTaskReloading<D> | AsyncTaskSuccessful<D> {
  return task.status === "reloading" || task.status === "successful";
}

export function isAsyncTaskStarting<D, E>(
  task: AsyncTask<D, E>
): task is AsyncTaskLoading | AsyncTaskPending {
  return task.status === "loading" || task.status === "pending";
}

export function hasAsyncTaskFailed<D, E>(task: AsyncTask<D, E>): task is AsyncTaskFailed<E> {
  return task.status === "failed";
}

export type Exact<T, U> = [T, U] extends [U, T] ? true : false;

export const StrictSchema: <Input, Output = Input>() => <ParserOutput, ParserInput = ParserOutput>(
  parser: Exact<
    ZodSchema<Output, ZodTypeDef, Input>,
    ZodSchema<ParserOutput, ZodTypeDef, ParserInput>
  > extends true
    ? Exact<Required<Output>, Required<ParserOutput>> extends true
      ? Exact<Required<Input>, Required<ParserInput>> extends true
        ? ZodSchema<ParserOutput, ZodTypeDef, ParserInput>
        : never
      : never
    : never
) => ZodSchema<Output, ZodTypeDef, Input> =
  () =>
  <Output, Input = Output>(parser: unknown) =>
    // eslint-disable-next-line no-type-assertion/no-type-assertion
    parser as ZodSchema<Output, ZodTypeDef, Input>;
