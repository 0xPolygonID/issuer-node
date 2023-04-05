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
