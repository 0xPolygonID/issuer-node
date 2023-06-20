import { z } from "zod";

import { getStrictParser } from "src/adapters/parsers";
import {
  ArrayAttribute,
  ArrayAttributeValue,
  Attribute,
  AttributeValue,
  BooleanAttribute,
  BooleanAttributeValue,
  BooleanProps,
  BooleanSchema,
  CommonProps,
  IntegerAttribute,
  IntegerAttributeValue,
  IntegerProps,
  IntegerSchema,
  JsonLdType,
  JsonSchema,
  JsonSchemaProps,
  MultiAttribute,
  MultiAttributeValue,
  MultiValue,
  NullAttribute,
  NullAttributeValue,
  NullSchema,
  NumberAttribute,
  NumberAttributeValue,
  NumberProps,
  NumberSchema,
  ObjectAttribute,
  ObjectAttributeValue,
  Schema,
  StringAttribute,
  StringAttributeValue,
  StringProps,
  StringSchema,
} from "src/domain";
import { extractCredentialSubjectAttributeWithoutId } from "src/utils/jsonSchemas";
import { Nullable } from "src/utils/types";

const commonPropsParser = getStrictParser<CommonProps>()(
  z.object({
    const: z.unknown().optional(),
    default: z.unknown().optional(),
    description: z.string().optional(),
    examples: z.unknown().array().optional(),
    title: z.string().optional(),
  })
);

// Primitives

const booleanPropsParser = getStrictParser<BooleanProps>()(
  z.object({
    enum: z.boolean().array().min(1).optional(),
  })
);

const booleanSchemaParser = getStrictParser<BooleanSchema>()(
  commonPropsParser.and(booleanPropsParser).and(
    z.object({
      type: z.literal("boolean"),
    })
  )
);

function getBooleanAttributeParser(name: string, required: boolean) {
  return getStrictParser<BooleanSchema, BooleanAttribute>()(
    booleanSchemaParser.transform(
      (schema): BooleanAttribute => ({
        name,
        required,
        schema,
        type: "boolean",
      })
    )
  );
}

const integerPropsParser = getStrictParser<IntegerProps>()(
  z.object({
    enum: z.number().array().min(1).optional(),
  })
);

const integerSchemaParser = getStrictParser<IntegerSchema>()(
  commonPropsParser.and(integerPropsParser).and(
    z.object({
      type: z.literal("integer"),
    })
  )
);

function getIntegerAttributeParser(name: string, required: boolean) {
  return getStrictParser<IntegerSchema, IntegerAttribute>()(
    integerSchemaParser.transform(
      (schema): IntegerAttribute => ({
        name,
        required,
        schema,
        type: "integer",
      })
    )
  );
}

const nullSchemaParser = getStrictParser<NullSchema>()(
  commonPropsParser.and(
    z.object({
      type: z.literal("null"),
    })
  )
);

function getNullAttributeParser(name: string, required: boolean) {
  return getStrictParser<NullSchema, NullAttribute>()(
    nullSchemaParser.transform(
      (schema): NullAttribute => ({
        name,
        required,
        schema,
        type: "null",
      })
    )
  );
}

const numberPropsParser = getStrictParser<NumberProps>()(
  z.object({
    enum: z.number().array().min(1).optional(),
  })
);

const numberSchemaParser = getStrictParser<NumberSchema>()(
  commonPropsParser.and(numberPropsParser).and(
    z.object({
      type: z.literal("number"),
    })
  )
);

function getNumberAttributeParser(name: string, required: boolean) {
  return getStrictParser<NumberSchema, NumberAttribute>()(
    numberSchemaParser.transform(
      (schema): NumberAttribute => ({
        name,
        required,
        schema,
        type: "number",
      })
    )
  );
}

const stringPropsParser = getStrictParser<StringProps>()(
  z.object({
    enum: z.string().array().min(1).optional(),
    format: z.string().optional(),
  })
);

const stringSchemaParser = getStrictParser<StringSchema>()(
  commonPropsParser.and(stringPropsParser).and(
    z.object({
      type: z.literal("string"),
    })
  )
);

function getStringAttributeParser(name: string, required: boolean) {
  return getStrictParser<StringSchema, StringAttribute>()(
    stringSchemaParser.transform(
      (schema): StringAttribute => ({
        name,
        required,
        schema,
        type: "string",
      })
    )
  );
}

// Non-primitive

type ArrayPropsInput = {
  items?: unknown;
};

type ArraySchemaInput = CommonProps & ArrayPropsInput & { type: "array" };

const arrayPropsInputParser = getStrictParser<ArrayPropsInput>()(
  z.object({
    items: z.unknown().optional(),
  })
);

const arraySchemaInputParser = getStrictParser<ArraySchemaInput>()(
  commonPropsParser.and(arrayPropsInputParser).and(
    z.object({
      type: z.literal("array"),
    })
  )
);

function getArrayAttributeParser(name: string, required: boolean) {
  return getStrictParser<ArraySchemaInput, ArrayAttribute>()(
    arraySchemaInputParser.transform(
      (schema, context): ArrayAttribute => ({
        name,
        required,
        schema: {
          ...schema,
          attribute: schema.items
            ? (() => {
                const parsed = getAttributeParser("items", required).safeParse(schema.items);
                if (parsed.success) {
                  return parsed.data;
                } else {
                  parsed.error.issues.map(context.addIssue);
                  return z.NEVER;
                }
              })()
            : undefined,
        },
        type: "array",
      })
    )
  );
}

type ObjectPropsInput = {
  properties?: Record<string, unknown>;
  required?: string[];
};

type ObjectSchemaInput = CommonProps & ObjectPropsInput & { type: "object" };

const objectPropsInputParser = getStrictParser<ObjectPropsInput>()(
  z.object({
    properties: z.record(z.unknown()).optional(),
    required: z.string().array().optional(),
  })
);

const objectSchemaInputParser = getStrictParser<ObjectSchemaInput>()(
  commonPropsParser.and(objectPropsInputParser).and(
    z.object({
      type: z.literal("object"),
    })
  )
);

function getObjectAttributeParser(name: string, required: boolean) {
  return getStrictParser<ObjectSchemaInput, ObjectAttribute>()(
    objectSchemaInputParser.transform(
      (schema, context): ObjectAttribute => ({
        name,
        required,
        schema: {
          ...schema,
          attributes:
            schema.properties &&
            Object.entries(schema.properties)
              .map(([name, value]) => {
                const parsed = getAttributeParser(
                  name,
                  schema.required !== undefined && schema.required.includes(name)
                ).safeParse(value);
                if (parsed.success) {
                  return parsed.data;
                } else {
                  parsed.error.issues.map(context.addIssue);
                  return z.NEVER;
                }
              })
              .sort((a, b) =>
                a.type !== "object" && b.type !== "object" ? 0 : a.type === "object" ? 1 : -1
              ),
        },
        type: "object",
      })
    )
  );
}

// Multi-type

type MultiSchemaType = {
  type: (
    | BooleanSchema["type"]
    | IntegerSchema["type"]
    | NullSchema["type"]
    | NumberSchema["type"]
    | StringSchema["type"]
    | ArraySchemaInput["type"]
    | ObjectSchemaInput["type"]
  )[];
};

type MultiProps = StringProps &
  IntegerProps &
  NumberProps &
  BooleanProps &
  ObjectPropsInput &
  ArrayPropsInput;

type MultiSchemaInput = CommonProps & MultiProps & MultiSchemaType;

const multiPropsParser = getStrictParser<MultiProps>()(
  stringPropsParser
    .and(integerPropsParser)
    .and(numberPropsParser)
    .and(booleanPropsParser)
    .and(objectPropsInputParser)
    .and(arrayPropsInputParser)
);

const multiSchemaTypeParser = getStrictParser<MultiSchemaType>()(
  z.object({
    type: z
      .union([
        z.literal("string"),
        z.literal("integer"),
        z.literal("number"),
        z.literal("boolean"),
        z.literal("null"),
        z.literal("object"),
        z.literal("array"),
      ])
      .array()
      .min(2),
  })
);

function getMultiAttributeParser(name: string, required: boolean) {
  return getStrictParser<MultiSchemaInput, MultiAttribute>()(
    commonPropsParser
      .and(multiPropsParser)
      .and(multiSchemaTypeParser)
      .transform(
        (schema, context): MultiAttribute => ({
          name,
          required,
          schemas: schema.type.map((type): Schema => {
            switch (type) {
              case "boolean": {
                const parsed = getBooleanAttributeParser(name, required).safeParse({
                  ...schema,
                  type: "boolean",
                });
                if (parsed.success) {
                  return parsed.data.schema;
                } else {
                  parsed.error.issues.map(context.addIssue);
                  return z.NEVER;
                }
              }
              case "integer": {
                const parsed = getIntegerAttributeParser(name, required).safeParse({
                  ...schema,
                  type: "integer",
                });
                if (parsed.success) {
                  return parsed.data.schema;
                } else {
                  parsed.error.issues.map(context.addIssue);
                  return z.NEVER;
                }
              }
              case "null": {
                const parsed = getNullAttributeParser(name, required).safeParse({
                  ...schema,
                  type: "null",
                });
                if (parsed.success) {
                  return parsed.data.schema;
                } else {
                  parsed.error.issues.map(context.addIssue);
                  return z.NEVER;
                }
              }
              case "number": {
                const parsed = getNumberAttributeParser(name, required).safeParse({
                  ...schema,
                  type: "number",
                });
                if (parsed.success) {
                  return parsed.data.schema;
                } else {
                  parsed.error.issues.map(context.addIssue);
                  return z.NEVER;
                }
              }
              case "string": {
                const parsed = getStringAttributeParser(name, required).safeParse({
                  ...schema,
                  type: "string",
                });
                if (parsed.success) {
                  return parsed.data.schema;
                } else {
                  parsed.error.issues.map(context.addIssue);
                  return z.NEVER;
                }
              }
              case "array": {
                const parsed = getArrayAttributeParser(name, required).safeParse({
                  ...schema,
                  type: "array",
                });
                if (parsed.success) {
                  return parsed.data.schema;
                } else {
                  parsed.error.issues.map(context.addIssue);
                  return z.NEVER;
                }
              }
              case "object": {
                const parsed = getObjectAttributeParser(name, required).safeParse({
                  ...schema,
                  type: "object",
                });
                if (parsed.success) {
                  return parsed.data.schema;
                } else {
                  parsed.error.issues.map(context.addIssue);
                  return z.NEVER;
                }
              }
            }
          }),
          type: "multi",
        })
      )
  );
}

// Schema

type SchemaInput =
  | BooleanSchema
  | IntegerSchema
  | NullSchema
  | NumberSchema
  | StringSchema
  | ArraySchemaInput
  | ObjectSchemaInput
  | MultiSchemaInput;

function getAttributeParser(name: string, required: boolean) {
  return getStrictParser<SchemaInput, Attribute>()(
    z.union([
      getBooleanAttributeParser(name, required),
      getIntegerAttributeParser(name, required),
      getNullAttributeParser(name, required),
      getNumberAttributeParser(name, required),
      getStringAttributeParser(name, required),
      getArrayAttributeParser(name, required),
      getObjectAttributeParser(name, required),
      getMultiAttributeParser(name, required),
    ])
  );
}

// Schema

const jsonSchemaPropsParser = getStrictParser<JsonSchemaProps>()(
  z.object({
    $metadata: z.object({
      uris: z.object({
        jsonLdContext: z.string(),
      }),
    }),
    $schema: z.string(),
  })
);

export const jsonSchemaParser = getStrictParser<JsonSchemaProps & ObjectSchemaInput, JsonSchema>()(
  jsonSchemaPropsParser
    .and(getObjectAttributeParser("schema", false))
    .transform(({ $metadata, $schema, ...rest }) => {
      const jsonSchemaProps: JsonSchemaProps = { $metadata, $schema };
      const attribute: ObjectAttribute = rest;

      return {
        ...attribute,
        jsonSchemaProps,
      };
    })
);
// JSON LD Type

export function getJsonLdTypeParser(jsonSchema: JsonSchema) {
  return getStrictParser<
    {
      "@context": [Record<string, unknown>];
    },
    JsonLdType[]
  >()(
    z
      .object({
        "@context": z.tuple([z.record(z.unknown())]),
      })
      .transform((ldContext, zodContext): JsonLdType[] => {
        const credentialSubjectAttribute = extractCredentialSubjectAttributeWithoutId(jsonSchema);

        if (!credentialSubjectAttribute) {
          zodContext.addIssue({
            code: z.ZodIssueCode.custom,
            fatal: true,
            message: "Couldn't find the attribute credentialSubject in the JSON Schema",
          });
          return z.NEVER;
        }

        const ldContextTypeParseResult = Object.entries(ldContext["@context"][0]).reduce(
          (acc: { success: false } | { success: true; value: JsonLdType[] }, [key, value]) => {
            const parsedValue = z
              .object({
                "@context": z.record(z.unknown()),
                "@id": z.string().url("Property @id of the type is not valid"),
              })
              .safeParse(value);

            const ldContextTypePropsParseResult = parsedValue.success
              ? credentialSubjectAttribute.schema.attributes?.reduce(
                  (acc: { success: true } | { error: string; success: false }, attribute) =>
                    acc.success && attribute.name in parsedValue.data["@context"]
                      ? acc
                      : {
                          error: `Couldn't find Property "${attribute.name}" of the JSON schema in the context`,
                          success: false,
                        },
                  { success: true }
                ) || {
                  error: "Couldn't find any properties in schema's credentialSubject",
                  success: false,
                }
              : { error: parsedValue.error.message, success: false };

            return parsedValue.success && ldContextTypePropsParseResult.success
              ? {
                  success: true,
                  value: acc.success
                    ? [...acc.value, { id: parsedValue.data["@id"], name: key }]
                    : [{ id: parsedValue.data["@id"], name: key }],
                }
              : acc;
          },
          { success: false }
        );

        if (ldContextTypeParseResult.success) {
          return ldContextTypeParseResult.value;
        } else {
          zodContext.addIssue({
            code: z.ZodIssueCode.custom,
            fatal: true,
            message: "Couldn't find any valid type in the JSON LD context of the schema",
          });
          return z.NEVER;
        }
      })
  );
}

// Values

function getBooleanAttributeValueParser({ name, required, schema, type }: BooleanAttribute) {
  return required
    ? getStrictParser<boolean, BooleanAttributeValue>()(
        z.boolean().transform(
          (value): BooleanAttributeValue => ({
            name,
            required,
            schema,
            type,
            value,
          })
        )
      )
    : getStrictParser<Nullable<boolean>, BooleanAttributeValue>()(
        z
          .boolean()
          .nullish()
          .transform((value): BooleanAttributeValue => {
            return {
              name,
              required,
              schema,
              type,
              value: value === null ? undefined : value,
            };
          })
      );
}

function getIntegerAttributeValueParser({ name, required, schema, type }: IntegerAttribute) {
  return required
    ? getStrictParser<number, IntegerAttributeValue>()(
        z.number().transform(
          (value): IntegerAttributeValue => ({
            name,
            required,
            schema,
            type,
            value,
          })
        )
      )
    : getStrictParser<Nullable<number>, IntegerAttributeValue>()(
        z
          .number()
          .nullish()
          .transform(
            (value): IntegerAttributeValue => ({
              name,
              required,
              schema,
              type,
              value: value === null ? undefined : value,
            })
          )
      );
}

function getNullAttributeValueParser({ name, required, schema, type }: NullAttribute) {
  return required
    ? getStrictParser<null, NullAttributeValue>()(
        z.null().transform(
          (value): NullAttributeValue => ({
            name,
            required,
            schema,
            type,
            value,
          })
        )
      )
    : getStrictParser<null | undefined, NullAttributeValue>()(
        z
          .null()
          .optional()
          .transform(
            (value): NullAttributeValue => ({
              name,
              required,
              schema,
              type,
              value,
            })
          )
      );
}

function getNumberAttributeValueParser({ name, required, schema, type }: NumberAttribute) {
  return required
    ? getStrictParser<number, NumberAttributeValue>()(
        z.number().transform(
          (value): NumberAttributeValue => ({
            name,
            required,
            schema,
            type,
            value,
          })
        )
      )
    : getStrictParser<Nullable<number>, NumberAttributeValue>()(
        z
          .number()
          .nullish()
          .transform(
            (value): NumberAttributeValue => ({
              name,
              required,
              schema,
              type,
              value: value === null ? undefined : value,
            })
          )
      );
}

function getStringAttributeValueParser({ name, required, schema, type }: StringAttribute) {
  return required
    ? getStrictParser<string, StringAttributeValue>()(
        z.string().transform(
          (value): StringAttributeValue => ({
            name,
            required,
            schema,
            type,
            value,
          })
        )
      )
    : getStrictParser<Nullable<string>, StringAttributeValue>()(
        z
          .string()
          .nullish()
          .transform(
            (value): StringAttributeValue => ({
              name,
              required,
              schema,
              type,
              value: value === null ? undefined : value,
            })
          )
      );
}

function getArrayAttributeValueParser({ name, required, schema, type }: ArrayAttribute) {
  const attribute = schema.attribute;
  return required
    ? getStrictParser<unknown[], ArrayAttributeValue>()(
        z.array(z.unknown()).transform(
          (unknowns, context): ArrayAttributeValue => ({
            name,
            required,
            schema,
            type,
            value: attribute
              ? unknowns.map((unknown) => {
                  const parsed = getAttributeValueParser(attribute).safeParse(unknown);
                  if (parsed.success) {
                    return parsed.data;
                  } else {
                    parsed.error.issues.map(context.addIssue);
                    return z.NEVER;
                  }
                })
              : [],
          })
        )
      )
    : getStrictParser<Nullable<unknown[]>, ArrayAttributeValue>()(
        z
          .array(z.unknown())
          .nullish()
          .transform(
            (unknowns, context): ArrayAttributeValue => ({
              name,
              required,
              schema,
              type,
              value: attribute
                ? unknowns === null
                  ? undefined
                  : unknowns &&
                    unknowns.map((unknown) => {
                      const parsed = getAttributeValueParser(attribute).safeParse(unknown);
                      if (parsed.success) {
                        return parsed.data;
                      } else {
                        parsed.error.issues.map(context.addIssue);
                        return z.NEVER;
                      }
                    })
                : [],
            })
          )
      );
}

function objectToObjectAttributeValue({
  context,
  object,
  objectAttribute,
}: {
  context: z.RefinementCtx;
  object: Record<string, unknown>;
  objectAttribute: ObjectAttribute;
}): ObjectAttributeValue {
  const { name, required, schema, type } = objectAttribute;

  // make sure all required properties of the objectAttribute are present in the object
  objectAttribute.schema.attributes?.forEach((attribute) => {
    const missing = attribute.required && Object.keys(object).includes(attribute.name) === false;
    if (missing) {
      context.addIssue({
        code: z.ZodIssueCode.custom,
        fatal: true,
        message: `Could not find the schema's required property "${attribute.name}" in the attribute "${name}".`,
      });
    }
  });

  return {
    name,
    required,
    schema,
    type,
    value: Object.entries(object)
      .reduce((acc: AttributeValue[], [name, unknown]) => {
        const attribute = schema.attributes?.find((attribute) => attribute.name === name);
        if (attribute) {
          const parsedAttributeValue = getAttributeValueParser(attribute).safeParse(unknown);
          if (parsedAttributeValue.success) {
            return [...acc, parsedAttributeValue.data];
          } else {
            parsedAttributeValue.error.issues.map((issue) => {
              context.addIssue({ ...issue, path: [attribute.name, ...issue.path] });
            });
          }
        }
        return acc;
      }, [])
      .sort((a, b) =>
        a.type !== "object" && b.type !== "object" ? 0 : a.type === "object" ? 1 : -1
      ),
  };
}

function getObjectAttributeValueParser(objectAttribute: ObjectAttribute) {
  const { name, required, schema, type } = objectAttribute;
  return required
    ? getStrictParser<Record<string, unknown>, ObjectAttributeValue>()(
        z
          .record(z.unknown())
          .transform(
            (object, context): ObjectAttributeValue =>
              objectToObjectAttributeValue({ context, object, objectAttribute })
          )
      )
    : getStrictParser<Nullable<Record<string, unknown>>, ObjectAttributeValue>()(
        z
          .record(z.unknown())
          .nullish()
          .transform(
            (object, context): ObjectAttributeValue =>
              object
                ? objectToObjectAttributeValue({ context, object, objectAttribute })
                : {
                    name,
                    required,
                    schema,
                    type,
                    value: undefined,
                  }
          )
      );
}

function parseMultiValue({
  name,
  required,
  schema,
  unknown,
}: {
  name: string;
  required: boolean;
  schema: Schema;
  unknown: unknown;
}) {
  switch (schema.type) {
    case "boolean": {
      return getBooleanAttributeValueParser({
        name,
        required,
        schema: schema,
        type: "boolean",
      }).safeParse(unknown);
    }
    case "integer": {
      return getIntegerAttributeValueParser({
        name,
        required,
        schema: schema,
        type: "integer",
      }).safeParse(unknown);
    }
    case "null": {
      return getNullAttributeValueParser({
        name,
        required,
        schema: schema,
        type: "null",
      }).safeParse(unknown);
    }
    case "number": {
      return getNumberAttributeValueParser({
        name,
        required,
        schema: schema,
        type: "number",
      }).safeParse(unknown);
    }
    case "string": {
      return getStringAttributeValueParser({
        name,
        required,
        schema: schema,
        type: "string",
      }).safeParse(unknown);
    }
    case "array": {
      return getArrayAttributeValueParser({
        name,
        required,
        schema: schema,
        type: "array",
      }).safeParse(unknown);
    }
    case "object": {
      return getObjectAttributeValueParser({
        name,
        required,
        schema: schema,
        type: "object",
      }).safeParse(unknown);
    }
  }
}

type MultiAttributeValueParseResult = { data: MultiValue; success: true } | { success: false };

function getMultiAttributeValueParser({ name, required, schemas, type }: MultiAttribute) {
  return getStrictParser<unknown, MultiAttributeValue>()(
    z.unknown().transform((unknown, context): MultiAttributeValue => {
      const value: MultiAttributeValueParseResult = schemas.reduce(
        (acc: MultiAttributeValueParseResult, schema: Schema): MultiAttributeValueParseResult => {
          if (acc.success) {
            return acc;
          } else {
            const result = parseMultiValue({
              name,
              required,
              schema,
              unknown,
            });
            return result.success ? result : acc;
          }
        },
        {
          success: false,
        }
      );

      if (value.success) {
        return {
          name,
          required,
          schemas,
          type,
          value: value.data,
        };
      } else {
        context.addIssue({
          code: z.ZodIssueCode.custom,
          fatal: true,
          message: `Could not parse the value of the multi attribute "${name}" against any of of its schemas.`,
        });
        return z.NEVER;
      }
    })
  );
}

export function getAttributeValueParser(attribute: Attribute) {
  return getStrictParser<unknown, AttributeValue>()(
    z.unknown().transform((unknown, context) => {
      switch (attribute.type) {
        case "boolean": {
          const parsed = getBooleanAttributeValueParser(attribute).safeParse(unknown);
          if (parsed.success) {
            return parsed.data;
          } else {
            parsed.error.issues.map(context.addIssue);
            return z.NEVER;
          }
        }
        case "integer": {
          const parsed = getIntegerAttributeValueParser(attribute).safeParse(unknown);
          if (parsed.success) {
            return parsed.data;
          } else {
            parsed.error.issues.map(context.addIssue);
            return z.NEVER;
          }
        }
        case "null": {
          const parsed = getNullAttributeValueParser(attribute).safeParse(unknown);
          if (parsed.success) {
            return parsed.data;
          } else {
            parsed.error.issues.map(context.addIssue);
            return z.NEVER;
          }
        }
        case "number": {
          const parsed = getNumberAttributeValueParser(attribute).safeParse(unknown);
          if (parsed.success) {
            return parsed.data;
          } else {
            parsed.error.issues.map(context.addIssue);
            return z.NEVER;
          }
        }
        case "string": {
          const parsed = getStringAttributeValueParser(attribute).safeParse(unknown);
          if (parsed.success) {
            return parsed.data;
          } else {
            parsed.error.issues.map(context.addIssue);
            return z.NEVER;
          }
        }
        case "array": {
          const parsed = getArrayAttributeValueParser(attribute).safeParse(unknown);
          if (parsed.success) {
            return parsed.data;
          } else {
            parsed.error.issues.map(context.addIssue);
            return z.NEVER;
          }
        }
        case "object": {
          const parsed = getObjectAttributeValueParser(attribute).safeParse(unknown);
          if (parsed.success) {
            return parsed.data;
          } else {
            parsed.error.issues.map(context.addIssue);
            return z.NEVER;
          }
        }
        case "multi": {
          const parsed = getMultiAttributeValueParser(attribute).safeParse(unknown);
          if (parsed.success) {
            return parsed.data;
          } else {
            parsed.error.issues.map(context.addIssue);
            return z.NEVER;
          }
        }
      }
    })
  );
}
