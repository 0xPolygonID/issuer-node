import { AxiosError } from "axios";
import { APIError } from "src/adapters/api";

export function isAbortedError(response: APIError | AxiosError) {
  return response.status === 0;
}

export function makeRequestAbortable<T>(request: (signal: AbortSignal) => T | Promise<T>) {
  const controller = new AbortController();

  return {
    aborter: () => controller.abort(),
    request: request(controller.signal),
  };
}
