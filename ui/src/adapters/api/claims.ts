import axios from "axios";
import { z } from "zod";

import { Schema, schema } from "src/adapters/api/schemas";
import {
  APIResponse,
  HTTPStatusSuccess,
  ResultCreated,
  ResultOK,
  buildAPIError,
} from "src/utils/adapters";
import {
  API_PASSWORD,
  API_URL,
  API_USERNAME,
  ISSUER_DID,
  QUERY_SEARCH_PARAM,
  VALID_SEARCH_PARAM,
} from "src/utils/constants";
import { StrictSchema } from "src/utils/types";

export interface ClaimAttribute {
  attributeKey: string;
  attributeValue: number;
}

export interface Claim {
  active: boolean;
  attributeValues: ClaimAttribute[];
  claimLinkExpiration: Date | null;
  createdAt: Date;
  expiresAt: Date | null;
  id: string;
  issuedClaims: number | null;
  limitedClaims: number | null;
  schemaTemplate: Schema;
  valid: boolean;
}

export interface ClaimIssuePayload {
  attributes: ClaimAttribute[];
  claimLinkExpiration?: string;
  expirationDate?: string;
  limitedClaims?: number;
}

export async function claimIssue({
  payload,
  schemaID,
}: {
  payload: ClaimIssuePayload;
  schemaID: string;
}): Promise<APIResponse<Claim>> {
  try {
    const response = await axios({
      baseURL: API_URL,
      data: payload,
      headers: {
        Authorization: `Basic ${API_USERNAME}:${API_PASSWORD}`,
      },
      method: "POST",
      url: `issuers/${ISSUER_DID}/schemas/${schemaID}/offers`,
    });
    const { data } = resultCreatedClaim.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

interface ClaimUpdatePayload {
  active: boolean;
}

export async function claimUpdate({
  claimID,
  payload,
  schemaID,
}: {
  claimID: string;
  payload: ClaimUpdatePayload;
  schemaID: string;
}): Promise<APIResponse<Claim>> {
  try {
    const response = await axios({
      baseURL: API_URL,
      data: payload,
      headers: {
        Authorization: `Basic ${API_USERNAME}:${API_PASSWORD}`,
      },
      method: "PATCH",
      url: `issuers/${ISSUER_DID}/schemas/${schemaID}/offers/${claimID}`,
    });
    const { data } = resultOKClaim.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

export async function claimsGetAll({
  params: { query, valid },
  signal,
}: {
  params: {
    query?: string;
    valid?: boolean;
  };
  signal?: AbortSignal;
}): Promise<
  APIResponse<{
    claims: Claim[];
    errors: z.ZodError<Claim>[];
  }>
> {
  try {
    const response = await axios({
      baseURL: API_URL,
      headers: {
        Authorization: `Basic ${API_USERNAME}:${API_PASSWORD}`,
      },
      method: "GET",
      params: new URLSearchParams({
        ...(query !== undefined ? { [QUERY_SEARCH_PARAM]: query } : {}),
        ...(valid !== undefined ? { [VALID_SEARCH_PARAM]: valid.toString() } : {}),
      }),
      signal,
      url: `issuers/${ISSUER_DID}/offers`,
    });
    const { data } = resultOKClaimsGetAll.parse(response);

    return {
      data: {
        claims: data.claims.sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime()),
        errors: data.errors,
      },
      isSuccessful: true,
    };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

export async function claimsGetSingle({
  claimID,
  schemaID,
  signal,
}: {
  claimID: string;
  schemaID: string;
  signal: AbortSignal;
}): Promise<APIResponse<Claim>> {
  try {
    const response = await axios({
      baseURL: API_URL,
      headers: {
        Authorization: `Basic ${API_USERNAME}:${API_PASSWORD}`,
      },
      method: "GET",
      signal,
      url: `issuers/${ISSUER_DID}/schemas/${schemaID}/offers/${claimID}`,
    });
    const { data } = resultOKClaim.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

interface ClaimQRCode {
  body: {
    callbackUrl: string;
    reason: string;
    scope: unknown[];
  };
  from: string;
  id: string;
  thid: string;
  typ: string;
  type: string;
}

const claimQRCode = StrictSchema<ClaimQRCode>()(
  z.object({
    body: z.object({
      callbackUrl: z.string(),
      reason: z.string(),
      scope: z.array(z.unknown()),
    }),
    from: z.string(),
    id: z.string(),
    thid: z.string(),
    typ: z.string(),
    type: z.string(),
  })
);

export interface ShareClaimQRCode {
  issuer: { displayName: string; logo: string };
  offerDetails: Claim;
  qrcode: ClaimQRCode;
  sessionID: string;
}

const apiClaimAttribute = StrictSchema<ClaimAttribute>()(
  z.object({
    attributeKey: z.string(),
    attributeValue: z.number(),
  })
);

const claim = StrictSchema<Claim>()(
  z.object({
    active: z.boolean(),
    attributeValues: z.array(apiClaimAttribute),
    claimLinkExpiration: z.coerce.date().nullable(),
    createdAt: z.coerce.date(),
    expiresAt: z.coerce.date().nullable(),
    id: z.string(),
    issuedClaims: z.number().nullable(),
    limitedClaims: z.number().nullable(),
    schemaTemplate: schema,
    valid: z.boolean(),
  })
);

export const shareClaimQRCode = StrictSchema<ShareClaimQRCode>()(
  z.object({
    issuer: z.object({
      displayName: z.string(),
      logo: z.string(),
    }),
    offerDetails: claim,
    qrcode: claimQRCode,
    sessionID: z.string(),
  })
);

const resultOKShareClaimQRCode = StrictSchema<ResultOK<ShareClaimQRCode>>()(
  z.object({
    data: shareClaimQRCode,
    status: z.literal(HTTPStatusSuccess.OK),
  })
);

export async function claimsQRCreate({
  id,
  signal,
}: {
  id: string;
  signal?: AbortSignal;
}): Promise<APIResponse<ShareClaimQRCode>> {
  try {
    const response = await axios({
      baseURL: API_URL,
      method: "POST",
      signal,
      url: `offers-qrcode/${id}`,
    });

    const { data } = resultOKShareClaimQRCode.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

const resultCreatedClaim = StrictSchema<ResultCreated<Claim>>()(
  z.object({
    data: claim,
    status: z.literal(HTTPStatusSuccess.Created),
  })
);

const resultOKClaim = StrictSchema<ResultOK<Claim>>()(
  z.object({
    data: claim,
    status: z.literal(HTTPStatusSuccess.OK),
  })
);

interface ClaimsGetAll {
  claims: Claim[];
  errors: z.ZodError<Claim>[];
}

const resultOKClaimsGetAll = StrictSchema<ResultOK<unknown[]>, ResultOK<ClaimsGetAll>>()(
  z.object({
    data: z.array(z.unknown()).transform((unknowns) =>
      unknowns.reduce(
        (acc: ClaimsGetAll, curr: unknown, index) => {
          const parsedClaim = claim.safeParse(curr);
          return parsedClaim.success
            ? {
                ...acc,
                claims: [...acc.claims, parsedClaim.data],
              }
            : {
                ...acc,
                errors: [
                  ...acc.errors,
                  new z.ZodError<Claim>(
                    parsedClaim.error.issues.map((issue) => ({
                      ...issue,
                      path: [index, ...issue.path],
                    }))
                  ),
                ],
              };
        },
        { claims: [], errors: [] }
      )
    ),
    status: z.literal(HTTPStatusSuccess.OK),
  })
);

interface AddingQRCode {
  body: {
    credentials: {
      description: string;
      id: string;
    }[];
    url: string;
  };
  from: string;
  id: string;
  thid: string;
  typ: string;
  type: string;
}

const addingQRCode = StrictSchema<AddingQRCode>()(
  z.object({
    body: z.object({
      credentials: z.array(
        z.object({
          description: z.string(),
          id: z.string(),
        })
      ),
      url: z.string(),
    }),
    from: z.string(),
    id: z.string(),
    thid: z.string(),
    typ: z.string(),
    type: z.string(),
  })
);

export enum ClaimQRStatus {
  Done = "done",
  Error = "error",
  Pending = "pending",
}

interface ClaimQRCheckDone {
  qrcode: AddingQRCode;
  status: ClaimQRStatus.Done;
}

interface ClaimQRCheckError {
  status: ClaimQRStatus.Error;
}

interface ClaimQRCheckPending {
  status: ClaimQRStatus.Pending;
}

export type ClaimQRCheck = ClaimQRCheckDone | ClaimQRCheckError | ClaimQRCheckPending;

const claimQRCheckDone = StrictSchema<ClaimQRCheckDone>()(
  z.object({
    qrcode: addingQRCode,
    status: z.literal(ClaimQRStatus.Done),
  })
);

const claimQRCheckError = StrictSchema<ClaimQRCheckError>()(
  z.object({
    status: z.literal(ClaimQRStatus.Error),
  })
);

const claimQRCheckPending = StrictSchema<ClaimQRCheckPending>()(
  z.object({
    status: z.literal(ClaimQRStatus.Pending),
  })
);

export const claimQRCheck = StrictSchema<ClaimQRCheck>()(
  z.union([claimQRCheckDone, claimQRCheckError, claimQRCheckPending])
);

const resultOKClaimQRCheck = StrictSchema<ResultOK<ClaimQRCheck>>()(
  z.object({
    data: claimQRCheck,
    status: z.literal(HTTPStatusSuccess.OK),
  })
);

export async function claimsQRCheck({
  claimID,
  sessionID,
}: {
  claimID: string;
  sessionID: string;
}): Promise<APIResponse<ClaimQRCheck>> {
  try {
    const response = await axios({
      baseURL: API_URL,
      method: "GET",
      params: {
        sessionID,
      },
      url: `offers-qrcode/${claimID}`,
    });

    const { data } = resultOKClaimQRCheck.parse(response);

    return { data, isSuccessful: true };
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}

export async function claimsQRDownload({
  claimID,
  sessionID,
}: {
  claimID: string;
  sessionID: string;
}): Promise<APIResponse<Blob>> {
  try {
    const response = await axios({
      baseURL: API_URL,
      method: "GET",
      params: {
        sessionID,
      },
      responseType: "blob",
      url: `offers-qrcode/${claimID}/download`,
    });

    if (response.data instanceof Blob) {
      return { data: response.data, isSuccessful: true };
    } else {
      return {
        error: { message: "Data returned by the API is not a valid file" },
        isSuccessful: false,
      };
    }
  } catch (error) {
    return { error: buildAPIError(error), isSuccessful: false };
  }
}
