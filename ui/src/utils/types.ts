import { AppError } from "src/domain/error";

export type ResourceMeta = {
  max_results: number;
  page: number;
  total: number;
};

export type List<T> = {
  failed: AppError[];
  successful: T[];
};

export type Resource<T> = {
  items: List<T>;
  meta: ResourceMeta;
};

export type Nullable<T> = T | null | undefined;
