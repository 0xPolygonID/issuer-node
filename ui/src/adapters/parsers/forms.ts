import dayjs, { isDayjs } from "dayjs";
import { z } from "zod";

import { CredentialAttribute } from "src/adapters/api/credentials";
import { SchemaAttribute } from "src/adapters/api/schemas";
import { CredentialForm, CredentialFormAttribute } from "src/domain/credentials";
import { ACCESSIBLE_UNTIL } from "src/utils/constants";
import { formatDate } from "src/utils/forms";
import { getStrictParser } from "src/utils/types";

// Types

interface LinkExpiration {
  linkExpirationDate?: dayjs.Dayjs | null;
  linkExpirationTime?: dayjs.Dayjs | null;
}

type IssueCredentialFormDataAttributesParsingResult =
  | {
      data: CredentialFormAttribute[];
      success: true;
    }
  | {
      error: z.ZodError;
      success: false;
    };

interface IssueCredentialFormDataInput {
  attributes: {
    attributes: Record<string, unknown>;
    expirationDate?: dayjs.Dayjs | null;
  };
  issuanceMethod: LinkExpiration & {
    linkMaximumIssuance?: number;
  };
}

// Parsers

const dayjsInstanceParser = getStrictParser<dayjs.Dayjs>()(
  z.custom<dayjs.Dayjs>(isDayjs, {
    message: "The provided input is not a valid Dayjs instance",
  })
);

const numericBooleanParser = getStrictParser<0 | 1, boolean>()(
  z.union([z.literal(0), z.literal(1)]).transform((value) => value === 1)
);

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

function parseIssueCredentialFormDataAttributes(
  attributes: Record<string, unknown>,
  schemaAttributes: SchemaAttribute[]
): IssueCredentialFormDataAttributesParsingResult {
  return Object.entries(attributes).reduce(
    (
      acc: IssueCredentialFormDataAttributesParsingResult,
      [attributeKey, attributeValue]: [string, unknown]
    ): IssueCredentialFormDataAttributesParsingResult => {
      if (!acc.success) {
        return acc;
      }

      const foundSchemaAttribute = schemaAttributes.find(
        (schemaAttribute) => schemaAttribute.name === attributeKey
      );

      if (!foundSchemaAttribute) {
        return {
          error: new z.ZodError([
            {
              code: z.ZodIssueCode.custom,
              message: `Could not find the attribute "${attributeKey}" in the schema attribute list`,
              path: [attributeKey],
            },
          ]),
          success: false,
        };
      } else {
        switch (foundSchemaAttribute.type) {
          case "date": {
            const parsedValue = dayjsInstanceParser.safeParse(attributeValue);

            return parsedValue.success
              ? {
                  data: [
                    ...acc.data,
                    {
                      name: attributeKey,
                      type: "date",
                      value: parsedValue.data.toDate(),
                    },
                  ],
                  success: true,
                }
              : {
                  error: parsedValue.error,
                  success: false,
                };
          }
          case "number": {
            const parsedValue = z.number().safeParse(attributeValue);

            return parsedValue.success
              ? {
                  data: [
                    ...acc.data,
                    {
                      name: attributeKey,
                      type: "number",
                      value: parsedValue.data,
                    },
                  ],
                  success: true,
                }
              : {
                  error: parsedValue.error,
                  success: false,
                };
          }
          case "boolean": {
            const parsedValue = numericBooleanParser.safeParse(attributeValue);

            return parsedValue.success
              ? {
                  data: [
                    ...acc.data,
                    {
                      name: attributeKey,
                      type: "boolean",
                      value: parsedValue.data,
                    },
                  ],
                  success: true,
                }
              : {
                  error: parsedValue.error,
                  success: false,
                };
          }
          case "singlechoice": {
            const parsedValue = z.number().safeParse(attributeValue);

            return parsedValue.success
              ? {
                  data: [
                    ...acc.data,
                    {
                      name: attributeKey,
                      type: "singlechoice",
                      value: parsedValue.data,
                    },
                  ],
                  success: true,
                }
              : {
                  error: parsedValue.error,
                  success: false,
                };
          }
        }
      }
    },
    { data: [], success: true }
  );
}

// Exports

export function getIssueCredentialFormDataParser(schemaAttributes: SchemaAttribute[]) {
  return getStrictParser<IssueCredentialFormDataInput, CredentialForm>()(
    z
      .object({
        attributes: z.object({
          attributes: z.record(z.unknown()),
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
            attributes: { attributes, expirationDate },
            issuanceMethod: { linkExpirationDate, linkExpirationTime, linkMaximumIssuance },
          },
          zodRefinementCtx
        ) => {
          const attributesParsingResult = parseIssueCredentialFormDataAttributes(
            attributes,
            schemaAttributes
          );

          if (attributesParsingResult.success) {
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
              attributes: attributesParsingResult.data,
              expiration: expirationDate ? expirationDate.toDate() : undefined,
              linkAccessibleUntil,
              linkMaximumIssuance,
            };
          } else {
            attributesParsingResult.error.issues.forEach(zodRefinementCtx.addIssue);

            return z.NEVER;
          }
        }
      )
  );
}

export function formatAttributeValue(
  attribute: CredentialAttribute,
  schemaAttributes: SchemaAttribute[]
):
  | {
      data: string;
      success: true;
    }
  | {
      error: string;
      success: false;
    } {
  const schemaAttribute = schemaAttributes.find((item) => item.name === attribute.attributeKey);

  if (!schemaAttribute) {
    return {
      error: "Not found",
      success: false,
    };
  } else {
    switch (schemaAttribute.type) {
      case "boolean": {
        const parsedBoolean = numericBooleanParser.safeParse(attribute.attributeValue);

        if (parsedBoolean.success) {
          return {
            data: parsedBoolean.data.toString(),
            success: true,
          };
        } else {
          return {
            error: `${attribute.attributeValue} cannot be parsed as boolean`,
            success: false,
          };
        }
      }

      case "date": {
        const momentInstance = dayjs(attribute.attributeValue.toString(), "YYYYMMDD");

        if (momentInstance.isValid()) {
          return {
            data: formatDate(momentInstance),
            success: true,
          };
        } else {
          return {
            error: "Date cannot be parsed",
            success: false,
          };
        }
      }

      case "number": {
        return {
          data: attribute.attributeValue.toString(),
          success: true,
        };
      }

      case "singlechoice": {
        const name = schemaAttribute.values.find(
          (choice) => choice.value === attribute.attributeValue
        );

        if (name) {
          return {
            data: `${name.name} (${attribute.attributeValue})`,
            success: true,
          };
        } else {
          return {
            error: `${attribute.attributeValue} cannot be parsed as single choice`,
            success: false,
          };
        }
      }
    }
  }
}

export const linkExpirationDateParser = getStrictParser<{
  linkExpirationDate: dayjs.Dayjs | null;
}>()(
  z.object({
    linkExpirationDate: dayjsInstanceParser.nullable(),
  })
);
