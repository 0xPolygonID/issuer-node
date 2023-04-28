import { z } from "zod";

export type List<T> = {
  failed: z.ZodError<T>[];
  successful: T[];
};
