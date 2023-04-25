import dayjs, { isDayjs } from "dayjs";
import { z } from "zod";

import { CreateLink } from "src/adapters/api/credentials";
import { jsonParser } from "src/adapters/json";
import { getStrictParser } from "src/adapters/parsers";
import { Json, ProofType } from "src/domain";
import { ACCESSIBLE_UNTIL } from "src/utils/constants";

// Types

type FormLiteral = string | number | boolean | null | undefined;
type FormLiteralInput = FormLiteral | dayjs.Dayjs;
type FormInput = { [key: string]: FormLiteralInput | FormInput };

type CredentialIssuance = {
  credentialSubject: Record<string, unknown> | undefined;
  expiration: Date | undefined;
  mtProof: boolean;
  signatureProof: boolean;
};

export type CredentialDirectIssuance = CredentialIssuance & {
  did: string;
  type: "directIssue";
};

export type CredentialLinkIssuance = CredentialIssuance & {
  linkAccessibleUntil: Date | undefined;
  linkMaximumIssuance: number | undefined;
  type: "credentialLink";
};

// Parsers

const dayjsInstanceParser = getStrictParser<dayjs.Dayjs>()(
  z.custom<dayjs.Dayjs>(isDayjs, {
    message: "The provided input is not a valid Dayjs instance",
  })
);

const formLiteralParser = getStrictParser<FormLiteralInput, FormLiteral>()(
  z.union([
    z.string(),
    z.number(),
    z.boolean(),
    z.null(),
    z.undefined(),
    dayjsInstanceParser.transform((dayjs) => dayjs.toISOString()),
  ])
);

const formParser: z.ZodType<Json, z.ZodTypeDef, FormInput> = getStrictParser<FormInput, Json>()(
  z
    .lazy(() => z.record(z.union([formLiteralParser, formParser])))
    .transform((data, context) => {
      const parsedJson = jsonParser.safeParse(data);
      if (parsedJson.success) {
        return parsedJson.data;
      } else {
        parsedJson.error.issues.map(context.addIssue);
        return z.NEVER;
      }
    })
);

interface LinkExpiration {
  linkExpirationDate?: dayjs.Dayjs | null;
  linkExpirationTime?: dayjs.Dayjs | null;
}

const linkExpirationParser = getStrictParser<LinkExpiration>()(
  z.object({
    linkExpirationDate: dayjsInstanceParser.nullable().optional(),
    linkExpirationTime: dayjsInstanceParser.nullable().optional(),
  })
);

export type IssuanceMethodFormData =
  | (LinkExpiration & {
      linkMaximumIssuance?: number;
      type: "credentialLink";
    })
  | {
      did: string;
      type: "directIssue";
    };

const issuanceMethodFormDataParser = getStrictParser<IssuanceMethodFormData>()(
  z.union([
    linkExpirationParser.and(
      z.object({
        linkMaximumIssuance: z.number().optional(),
        type: z.literal("credentialLink"),
      })
    ),
    z.object({
      did: z.string(),
      type: z.literal("directIssue"),
    }),
  ])
);

export type IssueCredentialFormData = {
  credentialSubject?: Record<string, unknown>;
  expirationDate?: dayjs.Dayjs | null;
  proofTypes: ProofType[];
};

const issueCredentialFormDataParser = getStrictParser<IssueCredentialFormData>()(
  z.object({
    credentialSubject: z.record(z.unknown()).optional(),
    expirationDate: dayjsInstanceParser.nullable().optional(),
    proofTypes: z.array(z.union([z.literal("MTP"), z.literal("SIG")])).min(1),
  })
);

export interface CredentialFormInput {
  issuanceMethod: IssuanceMethodFormData;
  issueCredential: IssueCredentialFormData;
}

export const credentialFormParser = getStrictParser<
  CredentialFormInput,
  CredentialLinkIssuance | CredentialDirectIssuance
>()(
  z
    .object({
      issuanceMethod: issuanceMethodFormDataParser,
      issueCredential: issueCredentialFormDataParser,
    })
    .transform(({ issuanceMethod, issueCredential }, context) => {
      const { credentialSubject, expirationDate, proofTypes } = issueCredential;
      const { type } = issuanceMethod;

      const baseIssuance = {
        credentialSubject,
        expiration: expirationDate ? expirationDate.toDate() : undefined,
        mtProof: proofTypes.includes("MTP"),
        signatureProof: proofTypes.includes("SIG"),
      };

      if (type === "credentialLink") {
        const { linkExpirationDate, linkExpirationTime, linkMaximumIssuance } = issuanceMethod;
        const linkAccessibleUntil = buildLinkAccessibleUntil({
          linkExpirationDate,
          linkExpirationTime,
        });

        const now = new Date().getTime();

        if (linkAccessibleUntil && linkAccessibleUntil.getTime() < now) {
          context.addIssue({
            code: z.ZodIssueCode.custom,
            fatal: true,
            message: `${ACCESSIBLE_UNTIL} must be a date/time in the future.`,
          });

          return z.NEVER;
        }

        return {
          ...baseIssuance,
          linkAccessibleUntil,
          linkMaximumIssuance,
          type,
        };
      }

      // Direct issuance
      const { did } = issuanceMethod;

      return {
        ...baseIssuance,
        did,
        type,
      };
    })
);

export const linkExpirationDateParser = getStrictParser<{
  linkExpirationDate: dayjs.Dayjs | null;
}>()(
  z.object({
    linkExpirationDate: dayjsInstanceParser.nullable(),
  })
);

// Serializers

export function serializeCredentialLinkIssuance({
  issueCredential: {
    credentialSubject,
    expiration,
    linkAccessibleUntil,
    linkMaximumIssuance,
    mtProof,
    signatureProof,
  },
  schemaID,
}: {
  issueCredential: CredentialLinkIssuance;
  schemaID: string;
}): { data: CreateLink; success: true } | { error: z.ZodError<FormInput>; success: false } {
  const parsedCredentialSubject = formParser.safeParse(credentialSubject);
  if (parsedCredentialSubject.success) {
    const claimLinkExpiration = linkAccessibleUntil ? linkAccessibleUntil.toISOString() : null;
    const expirationDate = expiration ? dayjs(expiration).format("YYYY-MM-DD") : null;
    const limitedClaims = linkMaximumIssuance ?? null;

    return {
      data: {
        claimLinkExpiration,
        credentialSubject: parsedCredentialSubject.data,
        expirationDate,
        limitedClaims,
        mtProof,
        schemaID,
        signatureProof,
      },
      success: true,
    };
  } else {
    return parsedCredentialSubject;
  }
}

// Helpers

function buildLinkAccessibleUntil({ linkExpirationDate, linkExpirationTime }: LinkExpiration) {
  if (linkExpirationDate) {
    if (linkExpirationTime) {
      return dayjs(linkExpirationDate)
        .set("hour", linkExpirationTime.get("hour"))
        .set("minute", linkExpirationTime.get("minute"))
        .set("second", linkExpirationTime.get("second"))
        .toDate();
    } else {
      return dayjs(linkExpirationDate).endOf("day").toDate();
    }
  } else {
    if (linkExpirationTime) {
      return linkExpirationTime.toDate();
    } else {
      return undefined;
    }
  }
}
