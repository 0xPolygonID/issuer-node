import { z } from "zod";

import { getStrictParser } from "src/adapters/parsers";
import { Env } from "src/domain";

export type Sorter = { field: string; order: "ascend" | "descend" };

export const serializeSorters = (sorters: Sorter[]) =>
  sorters.map(({ field, order }) => `${order === "descend" ? "-" : ""}${field}`).join(",");

export const parseSorters = (value: string | null) =>
  value === null || value === ""
    ? []
    : value.split(",").map(
        (field): Sorter => ({
          field: field.replace(/^-/, ""),
          order: field.startsWith("-") ? "descend" : "ascend",
        })
      );

export type ID = {
  id: string;
};

export const IDParser = getStrictParser<ID>()(z.object({ id: z.string() }));

export type Message = {
  message: string;
};

export const messageParser = getStrictParser<Message>()(z.object({ message: z.string() }));

export function buildAuthorizationHeader(env: Env) {
  return `Basic ${window.btoa(`${env.api.username}:${env.api.password}`)}`;
}
