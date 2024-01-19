import { z } from "zod";

export type ResourceMeta = {
  max_results: number;
  page: number;
  total: number;
};

export type List<T> = {
  failed: z.ZodError<T>[];
  successful: T[];
};

export type Resource<T> = {
  items: List<T>;
  meta: ResourceMeta;
};

export type Nullable<T> = T | null | undefined;
