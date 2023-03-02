import axios from "axios";
import { decodeToken } from "react-jwt";
import { z } from "zod";

import { APIResponse, HTTPStatusSuccess, ResultOK, buildAPIError } from "src/utils/adapters";
import { API_URL } from "src/utils/constants";
import { StrictSchema } from "src/utils/types";

type UserRole = "MEMBER" | "OWNER";

export interface Account {
  email: string;
  id: string;
  organization: null | string;
  role: UserRole;
  verified: boolean;
}

interface AuthenticationToken {
  account: AccountToken;
  exp: number;
  iat: number;
  jti: string;
  nbf: number;
  sub: string;
}

export function parseAccount(token: string): Account {
  const { account, sub } = decodeToken<AuthenticationToken>(token) || {};

  if (account && sub) {
    return { ...account, id: sub };
  }

  return {
    email: "",
    id: "",
    organization: null,
    role: "OWNER",
    verified: false,
  };
}

type AccountToken = Omit<Account, "id">;

const accountToken = StrictSchema<AccountToken>()(
  z.object({
    email: z.string(),
    organization: z.union([z.null(), z.string()]),
    role: z.union([z.literal("MEMBER"), z.literal("OWNER")]),
    verified: z.boolean(),
  })
);

export interface Authentication {
  email: string;
  password: string;
}

interface Token {
  token: string;
}

const resultOKToken = StrictSchema<ResultOK<Token>>()(
  z.object({
    data: z.object({ token: z.string() }),
    status: z.literal(HTTPStatusSuccess.OK),
  })
);

export async function authenticationSignIn({
  payload,
}: {
  payload: Authentication;
}): Promise<APIResponse<Token>> {
  try {
    const response = await axios({
      baseURL: API_URL,
      data: payload,
      method: "POST",
      url: "orgs/sign-in",
    });
    const {
      data: { token },
    } = resultOKToken.parse(response);

    accountToken.parse(decodeToken<AuthenticationToken>(token)?.account);

    return { data: { token }, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}
