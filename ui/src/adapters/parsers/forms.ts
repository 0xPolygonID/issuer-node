import dayjs, { isDayjs } from "dayjs";
import { z } from "zod";

import { CreateLink } from "src/adapters/api/credentials";
import { jsonParser } from "src/adapters/json";
import { getStrictParser } from "src/adapters/parsers";
import { Json } from "src/domain";
import { ACCESSIBLE_UNTIL } from "src/utils/constants";

// Types

interface LinkExpiration {
  linkExpirationDate?: dayjs.Dayjs | null;
  linkExpirationTime?: dayjs.Dayjs | null;
}

interface CredentialFormInput {
  credentialForm: {
    credentialSubject: Record<string, unknown>;
    expirationDate?: dayjs.Dayjs | null;
  };
  issuanceMethod: LinkExpiration & {
    linkMaximumIssuance?: number;
  };
}

interface CredentialForm {
  credentialSubject: Record<string, unknown>;
  expiration: Date | undefined;
  linkAccessibleUntil: Date | undefined;
  linkMaximumIssuance: number | undefined;
}

type FormLiteralInput = FormLiteral | dayjs.Dayjs;
type FormLiteral = string | number | boolean | null | undefined;
type FormInput = { [key: string]: FormLiteralInput | FormInput };

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

export const credentialFormParser = getStrictParser<CredentialFormInput, CredentialForm>()(
  z
    .object({
      credentialForm: z.object({
        credentialSubject: z.record(z.unknown()),
        expirationDate: dayjsInstanceParser.nullable().optional(),
      }),
      issuanceMethod: z.object({
        linkExpirationDate: dayjsInstanceParser.nullable().optional(),
        linkExpirationTime: dayjsInstanceParser.nullable().optional(),
        linkMaximumIssuance: z.number().optional(),
      }),
    })
    .transform(
      (
        {
          credentialForm: { credentialSubject, expirationDate },
          issuanceMethod: { linkExpirationDate, linkExpirationTime, linkMaximumIssuance },
        },
        zodRefinementCtx
      ) => {
        const linkAccessibleUntil = buildLinkAccessibleUntil({
          linkExpirationDate,
          linkExpirationTime,
        });

        const now = new Date().getTime();

        if (linkAccessibleUntil && linkAccessibleUntil.getTime() < now) {
          zodRefinementCtx.addIssue({
            code: z.ZodIssueCode.custom,
            fatal: true,
            message: `${ACCESSIBLE_UNTIL} must be a date/time in the future.`,
          });

          return z.NEVER;
        }

        return {
          credentialSubject,
          expiration: expirationDate ? expirationDate.toDate() : undefined,
          linkAccessibleUntil,
          linkMaximumIssuance,
        };
      }
    )
);

export const linkExpirationDateParser = getStrictParser<{
  linkExpirationDate: dayjs.Dayjs | null;
}>()(
  z.object({
    linkExpirationDate: dayjsInstanceParser.nullable(),
  })
);

// Serializers

export function serializeCredentialForm({
  credentialForm: { credentialSubject, expiration, linkAccessibleUntil, linkMaximumIssuance },
  schemaID,
}: {
  credentialForm: CredentialForm;
  schemaID: string;
}): { data: CreateLink; success: true } | { error: z.ZodError<FormInput>; success: false } {
  const parsedCredentialSubject = formParser.safeParse(credentialSubject);
  if (parsedCredentialSubject.success) {
    const expirationDate = expiration ? dayjs(expiration).toISOString() : null;
    const claimLinkExpiration = linkAccessibleUntil ? linkAccessibleUntil.toISOString() : null;
    const limitedClaims = linkMaximumIssuance !== undefined ? linkMaximumIssuance : null;

    return {
      data: {
        claimLinkExpiration,
        credentialSubject: parsedCredentialSubject.data,
        expirationDate,
        limitedClaims,
        mtProof: false,
        schemaID,
        signatureProof: true,
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
