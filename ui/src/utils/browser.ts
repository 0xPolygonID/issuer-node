import { APIError, HTTPStatusError } from "src/adapters/api";

export function isAbortedError(response: APIError) {
  return HTTPStatusError.Aborted === response.status;
}

export function makeRequestAbortable<T>(request: (signal: AbortSignal) => Promise<T>) {
  const controller = new AbortController();

  return {
    aborter: () => controller.abort(),
    request: request(controller.signal),
  };
}

export function downloadFile(blob: Blob, filename: string): void {
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement("a");

  a.style.display = "none";
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  window.URL.revokeObjectURL(url);
}
