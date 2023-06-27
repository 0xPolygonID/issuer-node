export type AsyncTask<Data, Error> =
  | AsyncTaskFailed<Error>
  | AsyncTaskLoading
  | AsyncTaskPending
  | AsyncTaskReloading<Data>
  | AsyncTaskSuccessful<Data>;

type AsyncTaskFailed<Error> = {
  error: Error;
  status: "failed";
};

type AsyncTaskLoading = {
  status: "loading";
};

type AsyncTaskPending = {
  status: "pending";
};

type AsyncTaskReloading<Data> = {
  data: Data;
  status: "reloading";
};

type AsyncTaskSuccessful<Data> = {
  data: Data;
  status: "successful";
};

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
