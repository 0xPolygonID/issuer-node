import { z } from "zod";

export type List<T> = {
  failed: z.ZodError<T>[];
  successful: T[];
};

export type Nullable<T> = T | null | undefined;
