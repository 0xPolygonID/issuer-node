import z from "zod";

import { getStrictParser } from "src/adapters/parsers";
import { Env } from "src/domain";

export interface ID {
  id: string;
}

export const IDParser = getStrictParser<ID>()(z.object({ id: z.string() }));

export interface Message {
  message: string;
}

export const messageParser = getStrictParser<Message>()(z.object({ message: z.string() }));

export function buildAuthorizationHeader(env: Env) {
  return `Basic ${window.btoa(`${env.api.username}:${env.api.password}`)}`;
}
