import { z } from "zod";

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

export function getStorageByKey<Input, Output>({
  defaultValue,
  key,
  parser,
}: {
  defaultValue: Output;
  key: string;
  parser: z.ZodSchema<Output, z.ZodTypeDef, Input>;
}) {
  const value = localStorage.getItem(key);

  if (value === null) {
    return setStorageByKey({ key, value: defaultValue });
  } else {
    const parsed = parser.safeParse(value);

    if (parsed.success) {
      return parsed.data;
    } else {
      try {
        const parsedJson = parser.safeParse(JSON.parse(value));

        if (parsedJson.success) {
          return parsedJson.data;
        } else {
          return setStorageByKey({ key, value: defaultValue });
        }
      } catch (_) {
        return setStorageByKey({ key, value: defaultValue });
      }
    }
  }
}

export function setStorageByKey<T>({ key, value }: { key: string; value: T }) {
  const string = typeof value === "string" ? value : JSON.stringify(value);

  localStorage.setItem(key, string);

  return value;
}

export function downloadQRCanvas(canvas: HTMLCanvasElement, fileName: string) {
  const imageURL = canvas.toDataURL("image/png");

  const downloadLink = document.createElement("a");
  downloadLink.href = imageURL;
  downloadLink.download = `${fileName}.png`;

  document.body.appendChild(downloadLink);
  downloadLink.click();
  document.body.removeChild(downloadLink);
}
