import { z } from "zod";

export function getStorageByKey<I, O>({
  defaultValue,
  key,
  parser,
}: {
  defaultValue: O;
  key: string;
  parser: z.ZodSchema<O, z.ZodTypeDef, I>;
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
