import { AxiosError } from "axios";
import { APIError, HTTPStatusError } from "src/adapters/api";

export function isAbortedError(response: APIError | AxiosError) {
  return HTTPStatusError.Aborted === response.status;
}

export function makeRequestAbortable<T>(request: (signal: AbortSignal) => T | Promise<T>) {
  const controller = new AbortController();

  return {
    aborter: () => controller.abort(),
    request: request(controller.signal),
  };
}
