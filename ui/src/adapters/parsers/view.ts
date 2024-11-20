import dayjs, { isDayjs } from "dayjs";
import { z } from "zod";

import { Sorter } from "src/adapters/api";
import { CreateCredential, CreateLink } from "src/adapters/api/credentials";
import { jsonParser } from "src/adapters/json";
import { getStrictParser } from "src/adapters/parsers";
import { getAttributeValueParser } from "src/adapters/parsers/jsonSchemas";
import {
  Attribute,
  AttributeValue,
  CredentialStatusType,
  IdentityType,
  Json,
  JsonLiteral,
  JsonObject,
  Method,
  ObjectAttribute,
  ProofType,
} from "src/domain";
import { ACCESSIBLE_UNTIL } from "src/utils/constants";

// Types

type FormLiteral = string | number | boolean | null | undefined;
type FormLiteralInput = FormLiteral | dayjs.Dayjs;
type FormInput = { [key: string]: FormLiteralInput | FormInput };

type CredentialIssuance = {
  credentialDisplayMethod: string | undefined;
  credentialExpiration: Date | undefined;
  credentialRefreshService: string | undefined;
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

export type IdentityDetailsFormData = {
  displayName: string;
};

export type IdentityFormData = {
  blockchain: string;
  credentialStatusType: CredentialStatusType;
  displayName: string;
  method: Method;
  network: string;
  type: IdentityType;
};

export const identityFormDataParser = getStrictParser<IdentityFormData>()(
  z.object({
    blockchain: z.string(),
    credentialStatusType: z.nativeEnum(CredentialStatusType),
    displayName: z.string(),
    method: z.nativeEnum(Method),
    network: z.string(),
    type: z.nativeEnum(IdentityType),
  })
);

// Parsers
export type TableSorterInput = { field: string; order?: "ascend" | "descend" | undefined };

const tableSorterInputParser = getStrictParser<TableSorterInput>()(
  z.object({
    field: z.string(),
    order: z.union([z.literal("ascend"), z.literal("descend")]).optional(),
  })
);

export const tableSorterParser = getStrictParser<TableSorterInput | unknown[], Sorter[]>()(
  z.union([
    z
      .unknown()
      .array()
      .transform((unknowns): Sorter[] =>
        unknowns.reduce((acc: Sorter[], curr): Sorter[] => {
          const parsedSorter = tableSorterInputParser.safeParse(curr);
          return parsedSorter.success && parsedSorter.data.order !== undefined
            ? [
                ...acc,
                {
                  field: parsedSorter.data.field,
                  order: parsedSorter.data.order,
                },
              ]
            : acc;
        }, [])
      ),
    tableSorterInputParser.transform((sorter): Sorter[] =>
      sorter.order !== undefined
        ? [
            {
              field: sorter.field,
              order: sorter.order,
            },
          ]
        : []
    ),
  ])
);

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
    dayjsInstanceParser.transform((value) =>
      value.isValid() ? serializeDate(value, "date-time") : undefined
    ),
  ])
);

const schemaFormValuesParser: z.ZodType<Json, z.ZodTypeDef, FormInput> = getStrictParser<
  FormInput,
  Json
>()(
  z
    .lazy(() => z.record(z.union([formLiteralParser, schemaFormValuesParser])))
    .transform((data, context) => {
      try {
        const valueWithoutUndefined: unknown = JSON.parse(JSON.stringify(data));
        const parsedJson = jsonParser.safeParse(valueWithoutUndefined);
        if (parsedJson.success) {
          return parsedJson.data;
        } else {
          parsedJson.error.issues.map(context.addIssue);
          return z.NEVER;
        }
      } catch (error) {
        context.addIssue({
          code: z.ZodIssueCode.custom,
          message: "The provided input is not a valid JSON object",
        });
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
  displayMethod: { enabled: boolean; url: string };
  proofTypes: ProofType[];
  refreshService: { enabled: boolean; url: string };
  schemaID?: string;
};

const issueCredentialFormDataParser = getStrictParser<IssueCredentialFormData>()(
  z.object({
    credentialExpiration: dayjsInstanceParser.nullable().optional(),
    credentialSubject: z.record(z.unknown()).optional(),
    displayMethod: z.object({
      enabled: z.boolean(),
      url: z.union([z.string().url(), z.literal("")]),
    }),
    proofTypes: z.array(z.nativeEnum(ProofType)).min(1, "At least one proof type is required"),
    refreshService: z.object({
      enabled: z.boolean(),
      url: z.union([z.string().url(), z.literal("")]),
    }),
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
      const { credentialExpiration, credentialSubject, displayMethod, proofTypes, refreshService } =
        issueCredential;
      const { type } = issuanceMethod;

      const baseIssuance = {
        credentialDisplayMethod: displayMethod.enabled ? displayMethod.url : undefined,
        credentialExpiration: credentialExpiration ? credentialExpiration.toDate() : undefined,
        credentialRefreshService: refreshService.enabled ? refreshService.url : undefined,
        credentialSubject,
        mtProof: proofTypes.includes(ProofType.Iden3SparseMerkleTreeProof),
        signatureProof: proofTypes.includes(ProofType.BJJSignature2021),
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
    credentialDisplayMethod,
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
        credentialExpiration: credentialExpiration
          ? serializeDate(credentialExpiration, "date-time")
          : null,
        credentialSubject: serializedSchemaForm.data === undefined ? {} : serializedSchemaForm.data,
        displayMethod: credentialDisplayMethod
          ? {
              id: credentialDisplayMethod,
              type: "Iden3BasicDisplayMethodv2",
            }
          : null,
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
    credentialDisplayMethod,
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
        displayMethod: credentialDisplayMethod
          ? {
              id: credentialDisplayMethod,
              type: "Iden3BasicDisplayMethodv2",
            }
          : null,
        expiration: credentialExpiration ? dayjs(credentialExpiration).unix() : null,
        proofs: [
          ...(mtProof ? [ProofType.Iden3SparseMerkleTreeProof] : []),
          ...(signatureProof ? [ProofType.BJJSignature2021] : []),
        ],
        refreshService: credentialRefreshService
          ? {
              id: credentialRefreshService,
              type: "Iden3RefreshService2023",
            }
          : null,
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
