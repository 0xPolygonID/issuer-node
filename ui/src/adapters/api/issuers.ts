import axios from "axios";
import { z } from "zod";

import { APIResponse, HTTPStatusSuccess, ResultOK, buildAPIError } from "src/utils/adapters";
import { API_URL } from "src/utils/constants";
import { StrictSchema } from "src/utils/types";

export interface ResponseIssuer {
  createdAt: Date;
  did: string;
  displayName: string;
  id: string;
  legalName: string | null;
  logo: string;
  modifiedAt: Date;
  ownerEmail: string;
  region: string;
  slug: string;
}

export const issuer = StrictSchema<ResponseIssuer>()(
  z.object({
    createdAt: z.coerce.date(),
    did: z.string(),
    displayName: z.string(),
    id: z.string(),
    legalName: z.string().nullable(),
    logo: z.string(),
    modifiedAt: z.coerce.date(),
    ownerEmail: z.string(),
    region: z.string(),
    slug: z.string(),
  })
);

export const resultOKResponseIssuer = StrictSchema<ResultOK<ResponseIssuer>>()(
  z.object({
    data: issuer,
    status: z.literal(HTTPStatusSuccess.OK),
  })
);

export async function issuersGet({
  id,
  signal,
  token,
}: {
  id: string;
  signal: AbortSignal;
  token: string;
}): Promise<APIResponse<ResponseIssuer>> {
  try {
    const response = await axios({
      baseURL: API_URL,
      headers: {
        Authorization: `Bearer ${token}`,
      },
      method: "GET",
      signal,
      url: `issuers/${id}`,
    });
    const { data } = resultOKResponseIssuer.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}
