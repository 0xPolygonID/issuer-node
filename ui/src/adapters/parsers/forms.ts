import dayjs, { isDayjs } from "dayjs";
import { z } from "zod";

import { PayloadValue } from "src/adapters";
import { CreateCredential, CreateLink } from "src/adapters/api/credentials";
import { jsonParser } from "src/adapters/json";
import { getStrictParser } from "src/adapters/parsers";
import { getAttributeValueParser } from "src/adapters/parsers/jsonSchemas";
import { Attribute, AttributeValue, Json, ObjectAttribute, ProofType } from "src/domain";
import { ACCESSIBLE_UNTIL } from "src/utils/constants";

// Types

type FormLiteralValue = string | number | boolean | null | undefined;
type FormLiteralInput = FormLiteralValue | dayjs.Dayjs;
type FormInput = { [key: string]: FormLiteralInput | FormInput };

type CredentialIssuance = {
  credentialExpiration: Date | undefined;
  credentialSubject: Record<string, unknown> | undefined;
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

const formLiteralParser = getStrictParser<FormLiteralInput, FormLiteralValue>()(
  z.union([
    z.string(),
    z.number(),
    z.boolean(),
    z.null(),
    z.undefined(),
    dayjsInstanceParser.transform((dayjs) => dayjs.toISOString()),
  ])
);

const schemaFormValuesParser: z.ZodType<Json, z.ZodTypeDef, FormInput> = getStrictParser<
  FormInput,
  Json
>()(
  z
    .lazy(() => z.record(z.union([formLiteralParser, schemaFormValuesParser])))
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
      did?: string;
      type: "directIssue";
    };

export const issuanceMethodFormDataParser = getStrictParser<IssuanceMethodFormData>()(
  z.union([
    linkExpirationParser.and(
      z.object({
        linkMaximumIssuance: z.number().optional(),
        type: z.literal("credentialLink"),
      })
    ),
    z.object({
      did: z.string().optional(),
      type: z.literal("directIssue"),
    }),
  ])
);

export type IssueCredentialFormData = {
  credentialExpiration?: dayjs.Dayjs | null;
  credentialSubject?: Record<string, unknown>;
  proofTypes: ProofType[];
};

const issueCredentialFormDataParser = getStrictParser<IssueCredentialFormData>()(
  z.object({
    credentialExpiration: dayjsInstanceParser.nullable().optional(),
    credentialSubject: z.record(z.unknown()).optional(),
    proofTypes: z
      .array(z.union([z.literal("MTP"), z.literal("SIG")]))
      .min(1, "At least one proof type is required"),
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
      const { credentialExpiration, credentialSubject, proofTypes } = issueCredential;
      const { type } = issuanceMethod;

      const baseIssuance = {
        credentialExpiration: credentialExpiration ? credentialExpiration.toDate() : undefined,
        credentialSubject,
        mtProof: proofTypes.includes("MTP"),
        signatureProof: proofTypes.includes("SIG"),
      };

      if (type === "credentialLink") {
        // Link issuance
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
      } else {
        // Direct issuance
        if (!issuanceMethod.did) {
          context.addIssue({
            code: z.ZodIssueCode.custom,
            fatal: true,
            message: `A connection or identifier must be provided.`,
          });
          return z.NEVER;
        }

        return {
          ...baseIssuance,
          did: issuanceMethod.did,
          type,
        };
      }
    })
);

// Serializers

function serializeDate(date: dayjs.Dayjs | Date, format: "date" | "date-time" | "time") {
  const template =
    format === "date"
      ? "YYYY-MM-DD"
      : format === "date-time"
      ? "YYYY-MM-DDTHH:mm:ss.SSSZ"
      : "HH:mm:ss.SSSZ";

  return dayjs(date).format(template);
}

function serializeAtrributeValue({
  attributeValue,
  init = {},
}: {
  attributeValue: AttributeValue;
  init?: PayloadValue;
}): PayloadValue[string] {
  switch (attributeValue.type) {
    case "integer":
    case "number":
    case "null":
    case "boolean": {
      return attributeValue.value === undefined ? init : attributeValue.value;
    }
    case "string": {
      const value = attributeValue.value;
      if (value === undefined) {
        return init;
      } else {
        const parsedDate = z.coerce.date(z.string().datetime()).safeParse(value);
        switch (attributeValue.schema.format) {
          case "date":
          case "date-time":
          case "time": {
            return parsedDate.success
              ? serializeDate(parsedDate.data, attributeValue.schema.format)
              : value;
          }
          default: {
            return value;
          }
        }
      }
    }
    case "object": {
      return attributeValue.value !== undefined
        ? attributeValue.value.reduce(
            (acc, curr) => ({
              ...acc,
              [curr.name]: serializeAtrributeValue({ attributeValue: curr }),
            }),
            {}
          )
        : init;
    }
    case "array": {
      return init;
    }
    case "multi": {
      return init;
    }
  }
}

export function serializeSchemaForm({
  attribute,
  ignoreRequired = false,
  value,
}: {
  attribute: Attribute;
  ignoreRequired?: boolean;
  value: Record<string, unknown>;
}): { data: PayloadValue[string]; success: true } | { error: z.ZodError; success: false } {
  const parsedSchemaFormValues = schemaFormValuesParser.safeParse(value);
  if (parsedSchemaFormValues.success) {
    const parsedAttributeValue = getAttributeValueParser(attribute, ignoreRequired).safeParse(
      parsedSchemaFormValues.data
    );
    if (parsedAttributeValue.success) {
      return {
        data: serializeAtrributeValue({ attributeValue: parsedAttributeValue.data }),
        success: true,
      };
    } else {
      return parsedAttributeValue;
    }
  } else {
    return parsedSchemaFormValues;
  }
}

export function serializeCredentialLinkIssuance({
  attribute,
  issueCredential: {
    credentialExpiration,
    credentialSubject,
    linkAccessibleUntil,
    linkMaximumIssuance,
    mtProof,
    signatureProof,
  },
  schemaID,
}: {
  attribute: ObjectAttribute;
  issueCredential: CredentialLinkIssuance;
  schemaID: string;
}): { data: CreateLink; success: true } | { error: z.ZodError<FormInput>; success: false } {
  const serializedSchemaForm = serializeSchemaForm({
    attribute,
    value: credentialSubject || {},
  });
  if (serializedSchemaForm.success) {
    return {
      data: {
        credentialExpiration: credentialExpiration
          ? serializeDate(credentialExpiration, "date")
          : null,
        credentialSubject: serializedSchemaForm.data,
        expiration: linkAccessibleUntil ? linkAccessibleUntil.toISOString() : null,
        limitedClaims: linkMaximumIssuance ?? null,
        mtProof,
        schemaID,
        signatureProof,
      },
      success: true,
    };
  } else {
    return serializedSchemaForm;
  }
}

export function serializeCredentialIssuance({
  attribute,
  credentialSchema,
  issueCredential: { credentialExpiration, credentialSubject, did, mtProof, signatureProof },
  type,
}: {
  attribute: ObjectAttribute;
  credentialSchema: string;
  issueCredential: CredentialDirectIssuance;
  type: string;
}): { data: CreateCredential; success: true } | { error: z.ZodError<FormInput>; success: false } {
  const serializedSchemaForm = serializeSchemaForm({
    attribute,
    value: {
      ...(credentialSubject || {}),
      id: did,
    },
  });
  if (serializedSchemaForm.success) {
    return {
      data: {
        credentialSchema,
        credentialSubject: serializedSchemaForm.data,
        expiration: credentialExpiration ? dayjs(credentialExpiration).toISOString() : null,
        mtProof,
        signatureProof,
        type,
      },
      success: true,
    };
  } else {
    return serializedSchemaForm;
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
