import { z } from "zod";

export type Meta = {
  current: number;
  max_results: number;
  total: number;
};

export type List<T> = {
  failed: z.ZodError<T>[];
  successful: T[];
};

export type Resource<T> = {
  items: List<T>;
  meta: Meta;
};

export type Nullable<T> = T | null | undefined;
