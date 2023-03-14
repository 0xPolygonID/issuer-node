import dayjs, { isDayjs } from "dayjs";
import { z } from "zod";

import { ClaimAttribute } from "src/adapters/api/claims";
import { SchemaAttribute } from "src/adapters/api/schemas";
import { ClaimForm, ClaimFormAttribute } from "src/domain";
import { numericBoolean } from "src/utils/adapters";
import { FORM_LABEL } from "src/utils/constants";
import { formatDate } from "src/utils/forms";
import { StrictSchema } from "src/utils/types";

const dayjsInstance = StrictSchema<dayjs.Dayjs>()(
  z.custom<dayjs.Dayjs>(isDayjs, {
    message: "The provided input is not a valid Dayjs instance",
  })
);

type IssueClaimFormDataAttributesParsingResult =
  | {
      data: ClaimFormAttribute[];
      success: true;
    }
  | {
      error: z.ZodError;
      success: false;
    };

const parseIssueClaimFormDataAttributes = (
  attributes: Record<string, unknown>,
  schemaAttributes: SchemaAttribute[]
): IssueClaimFormDataAttributesParsingResult => {
  return Object.entries(attributes).reduce(
    (
      acc: IssueClaimFormDataAttributesParsingResult,
      [attributeKey, attributeValue]: [string, unknown]
    ): IssueClaimFormDataAttributesParsingResult => {
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
            const parsedValue = dayjsInstance.safeParse(attributeValue);
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
            const parsedValue = numericBoolean.safeParse(attributeValue);
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
};

export const parseClaimLinkExpirationDate = StrictSchema<{
  claimLinkExpirationDate: dayjs.Dayjs | null;
}>()(
  z.object({
    claimLinkExpirationDate: dayjsInstance.nullable(),
  })
);

const issueClaimFormDataInput = z.object({
  attributes: z.object({
    attributes: z.record(z.unknown()),
    expirationDate: dayjsInstance.nullish(),
  }),
  issuanceMethod: z.object({
    claimLinkExpirationDate: dayjsInstance.nullish(),
    claimLinkExpirationTime: dayjsInstance.nullish(),
    limitedClaims: z.number().optional(),
  }),
});

type IssueClaimFormDataInput = z.infer<typeof issueClaimFormDataInput>;

export const issueClaimFormData = (schemaAttributes: SchemaAttribute[]) => {
  return StrictSchema<IssueClaimFormDataInput, ClaimForm>()(
    issueClaimFormDataInput.transform(
      (
        {
          attributes: { attributes, expirationDate },
          issuanceMethod: { claimLinkExpirationDate, claimLinkExpirationTime, limitedClaims },
        },
        zodRefinementCtx
      ) => {
        const attributesParsingResult = parseIssueClaimFormDataAttributes(
          attributes,
          schemaAttributes
        );

        if (attributesParsingResult.success) {
          const claimLinkExpiration = buildClaimLinkExpiration(
            claimLinkExpirationDate || undefined,
            claimLinkExpirationTime || undefined
          );

          const now = new Date().getTime();
          if (claimLinkExpiration && claimLinkExpiration.getTime() < now) {
            zodRefinementCtx.addIssue({
              code: z.ZodIssueCode.custom,
              fatal: true,
              message: `"${FORM_LABEL.LINK_VALIDITY}" must be a date/time in the future.`,
            });
            return z.NEVER;
          }

          return {
            attributes: attributesParsingResult.data,
            claimLinkExpiration,
            expirationDate: expirationDate ? expirationDate.toDate() : undefined,
            limitedClaims,
          };
        } else {
          attributesParsingResult.error.issues.forEach(zodRefinementCtx.addIssue);
          return z.NEVER;
        }
      }
    )
  );
};

const buildClaimLinkExpiration = (date?: dayjs.Dayjs, time?: dayjs.Dayjs): Date | undefined => {
  if (date) {
    if (time) {
      return dayjs(date)
        .set("hour", time.get("hour"))
        .set("minute", time.get("minute"))
        .set("second", time.get("second"))
        .toDate();
    } else {
      return dayjs(date).endOf("day").toDate();
    }
  } else {
    if (time) {
      return time.toDate();
    } else {
      return undefined;
    }
  }
};

export function formatAttributeValue(
  attribute: ClaimAttribute,
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

      case "boolean": {
        const parsedBoolean = numericBoolean.safeParse(attribute.attributeValue);

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
