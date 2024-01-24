import dayjs, { isDayjs } from "dayjs";
import { z } from "zod";

import { CreateCredential, CreateLink } from "src/adapters/api/credentials";
import { jsonParser } from "src/adapters/json";
import { getStrictParser } from "src/adapters/parsers";
import { getAttributeValueParser } from "src/adapters/parsers/jsonSchemas";
import {
  Attribute,
  AttributeValue,
  Json,
  JsonLiteral,
  JsonObject,
  ObjectAttribute,
  ProofType,
} from "src/domain";
import { ACCESSIBLE_UNTIL } from "src/utils/constants";

// Types

type FormLiteral = string | number | boolean | null | undefined;
type FormLiteralInput = FormLiteral | dayjs.Dayjs;
type FormInput = { [key: string]: FormLiteralInput | FormInput };

type CredentialIssuance = {
  credentialExpiration: Date | undefined;
  credentialRefreshService: string | null;
  credentialSubject: Record<string, unknown> | undefined;
  mtProof: boolean;
  signatureProof: boolean;
};

export type CredentialDirectIssuance = CredentialIssuance & {
  did: string;
  type: "directIssue";
};

export type CredentialLinkIssuance = CredentialIssuance & {
  credentialRefreshService: string | null;
  linkAccessibleUntil: Date | undefined;
  linkMaximumIssuance: number | undefined;
  type: "credentialLink";
};

// Parsers
export const dayjsInstanceParser = getStrictParser<dayjs.Dayjs>()(
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
    dayjsInstanceParser.transform((value) => (value.isValid() ? serializeDate(value, "date-time") : undefined)),
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

type LinkExpiration = {
  linkExpirationDate?: dayjs.Dayjs | null;
  linkExpirationTime?: dayjs.Dayjs | null;
};

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
  refreshService: { enabled: boolean; url: string };
  schemaID?: string;
};

const issueCredentialFormDataParser = getStrictParser<IssueCredentialFormData>()(
  z.object({
    credentialExpiration: dayjsInstanceParser.nullable().optional(),
    credentialSubject: z.record(z.unknown()).optional(),
    proofTypes: z
      .array(z.union([z.literal("MTP"), z.literal("SIG")]))
      .min(1, "At least one proof type is required"),
    refreshService: z.union([
      z.object({
        enabled: z.literal(false),
        url: z.string(),
      }),
      z.object({
        enabled: z.literal(true),
        url: z.string().url({
          message: `Refresh service URL must be a valid URL.`,
        }),
      }),
    ]),
    schemaID: z.string().optional(),
  })
);

export type CredentialFormInput = {
  issuanceMethod: IssuanceMethodFormData;
  issueCredential: IssueCredentialFormData;
};

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
      const { credentialExpiration, credentialSubject, proofTypes, refreshService } =
        issueCredential;
      const { type } = issuanceMethod;

      const baseIssuance = {
        credentialExpiration: credentialExpiration ? credentialExpiration.toDate() : undefined,
        credentialRefreshService: refreshService.enabled ? refreshService.url : null,
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

function serializeAttributeValue({
  attributeValue,
}: {
  attributeValue: AttributeValue;
}): JsonLiteral | JsonObject | undefined {
  switch (attributeValue.type) {
    case "integer":
    case "number": {
      const parsedConst = z.number().safeParse(attributeValue.schema.const);
      return parsedConst.success ? parsedConst.data : attributeValue.value;
    }
    case "boolean": {
      const parsedConst = z.boolean().safeParse(attributeValue.schema.const);
      return parsedConst.success ? parsedConst.data : attributeValue.value;
    }
    case "string": {
      const parsedConst = z.string().safeParse(attributeValue.schema.const);
      if (parsedConst.success) {
        return parsedConst.data;
      } else {
        switch (attributeValue.schema.format) {
          case "date":
          case "date-time":
          case "time": {
            const parsedDate = z.coerce.date().safeParse(attributeValue.value);
            return parsedDate.success
              ? serializeDate(parsedDate.data, attributeValue.schema.format)
              : attributeValue.value;
          }
          default: {
            return attributeValue.value;
          }
        }
      }
    }
    case "object": {
      return attributeValue.value !== undefined
        ? attributeValue.value.reduce((acc: JsonObject | undefined, curr) => {
            const value = serializeAttributeValue({ attributeValue: curr });
            return value !== undefined ? { ...acc, [curr.name]: value } : acc;
          }, undefined)
        : undefined;
    }
    case "null": {
      return attributeValue.value;
    }
    case "array": {
      return undefined;
    }
    case "multi": {
      return undefined;
    }
  }
}

export function serializeSchemaForm({
  attribute,
  value,
}: {
  attribute: Attribute;
  value: Record<string, unknown>;
}):
  | { data: JsonLiteral | JsonObject | undefined; success: true }
  | { error: z.ZodError; success: false } {
  const parsedSchemaFormValues = schemaFormValuesParser.safeParse(value);
  if (parsedSchemaFormValues.success) {
    const parsedAttributeValue = getAttributeValueParser(attribute).safeParse(
      parsedSchemaFormValues.data
    );
    if (parsedAttributeValue.success) {
      return {
        data: serializeAttributeValue({ attributeValue: parsedAttributeValue.data }),
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
    credentialRefreshService,
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
    value: credentialSubject === undefined ? {} : credentialSubject,
  });
  if (serializedSchemaForm.success) {
    return {
      data: {
        credentialExpiration: credentialExpiration ? serializeDate(credentialExpiration, "date-time") : null,
        credentialSubject: serializedSchemaForm.data === undefined ? {} : serializedSchemaForm.data,
        expiration: linkAccessibleUntil ? serializeDate(linkAccessibleUntil, "date-time") : null,
        limitedClaims: linkMaximumIssuance ?? null,
        mtProof,
        refreshService: credentialRefreshService
          ? {
              id: credentialRefreshService,
              type: "Iden3RefreshService2023",
            }
          : null,
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
  issueCredential: {
    credentialExpiration,
    credentialRefreshService,
    credentialSubject,
    did,
    mtProof,
    signatureProof,
  },
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
      ...(credentialSubject === undefined ? {} : credentialSubject),
      id: did,
    },
  });
  if (serializedSchemaForm.success) {
    return {
      data: {
        credentialSchema,
        credentialSubject: serializedSchemaForm.data === undefined ? {} : serializedSchemaForm.data,
        expiration: credentialExpiration ? serializeDate(dayjs(credentialExpiration), "date-time") : null,
        mtProof,
        refreshService: credentialRefreshService
          ? {
              id: credentialRefreshService,
              type: "Iden3RefreshService2023",
            }
          : null,
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
