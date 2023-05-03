import { AppError } from "src/domain";

export function isAbortedError(error: AppError) {
  return error.type === "cancel-error";
}

export function makeRequestAbortable<T>(request: (signal: AbortSignal) => T | Promise<T>) {
  const controller = new AbortController();

  return {
    aborter: () => controller.abort(),
    request: request(controller.signal),
  };
}
