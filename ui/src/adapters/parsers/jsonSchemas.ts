import { z } from "zod";

import { getStrictParser } from "src/adapters/parsers";
import {
  ArrayAttribute,
  Attribute,
  BooleanAttribute,
  BooleanProps,
  BooleanSchema,
  CommonProps,
  IntegerAttribute,
  IntegerSchema,
  JsonLdType,
  JsonSchema,
  MultiAttribute,
  MultiSchema,
  NullAttribute,
  NullSchema,
  NumberAttribute,
  NumberSchema,
  ObjectAttribute,
  ObjectProps,
  SchemaProps,
  StringAttribute,
  StringProps,
  StringSchema,
} from "src/domain";

const commonPropsParser = getStrictParser<CommonProps>()(
  z.object({
    description: z.string().optional(),
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

const integerSchemaParser = getStrictParser<IntegerSchema>()(
  commonPropsParser.and(
    z.object({
      enum: z.number().int().array().min(1).optional(),
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

const numberSchemaParser = getStrictParser<NumberSchema>()(
  commonPropsParser.and(
    z.object({
      enum: z.number().array().min(1).optional(),
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
  item?: unknown;
};

type ArraySchemaInput = CommonProps & ArrayPropsInput & { type: "array" };

const arrayPropsParser = getStrictParser<ArrayPropsInput>()(
  z.object({
    item: z.unknown().optional(),
  })
);

const arraySchemaParser = getStrictParser<ArraySchemaInput>()(
  commonPropsParser.and(arrayPropsParser).and(
    z.object({
      type: z.literal("array"),
    })
  )
);

function getArrayAttributeParser(name: string, required: boolean) {
  return getStrictParser<ArraySchemaInput, ArrayAttribute>()(
    arraySchemaParser.transform(
      (schema): ArrayAttribute => ({
        name,
        required,
        schema: {
          ...schema,
          item: schema.item ? getAttributeParser("items", required).parse(schema.item) : undefined,
        },
        type: "array",
      })
    )
  );
}

type ObjectPropsInput = Omit<ObjectProps, "properties"> & {
  properties?: Record<string, unknown>;
};

type ObjectSchemaInput = CommonProps & ObjectPropsInput & { type: "object" };

const objectPropsParser = getStrictParser<ObjectPropsInput>()(
  z.object({
    properties: z.record(z.unknown()).optional(),
    required: z.string().array().optional(),
  })
);

const objectSchemaParser = getStrictParser<ObjectSchemaInput>()(
  commonPropsParser.and(objectPropsParser).and(
    z.object({
      type: z.literal("object"),
    })
  )
);

function getObjectAttributeParser(name: string, required: boolean) {
  return getStrictParser<ObjectSchemaInput, ObjectAttribute>()(
    objectSchemaParser.transform(
      (schema): ObjectAttribute => ({
        name,
        required,
        schema: {
          ...schema,
          properties: schema.properties
            ? Object.entries(schema.properties).map(([name, value]) =>
                getAttributeParser(
                  name,
                  schema.required !== undefined && schema.required.includes(name)
                ).parse(value)
              )
            : undefined,
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

type MultiProps = StringProps & BooleanProps & ObjectPropsInput & ArrayPropsInput;

type MultiSchemaInput = CommonProps & MultiProps & MultiSchemaType;

const multiPropsParser = getStrictParser<MultiProps>()(
  stringPropsParser.and(booleanPropsParser).and(objectPropsParser).and(arrayPropsParser)
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

function getMultiSchemaParser(name: string, required: boolean) {
  return getStrictParser<MultiSchemaInput, MultiAttribute>()(
    commonPropsParser
      .and(multiPropsParser)
      .and(multiSchemaTypeParser)
      .transform(
        (schema): MultiAttribute => ({
          name,
          required,
          schemas: schema.type.map((type): MultiSchema[number] => {
            switch (type) {
              case "boolean": {
                return getBooleanAttributeParser(name, required).parse({
                  ...schema,
                  type: "boolean",
                }).schema;
              }
              case "integer": {
                return getIntegerAttributeParser(name, required).parse({
                  ...schema,
                  type: "integer",
                }).schema;
              }
              case "null": {
                return getNullAttributeParser(name, required).parse({ ...schema, type: "null" })
                  .schema;
              }
              case "number": {
                return getNumberAttributeParser(name, required).parse({ ...schema, type: "number" })
                  .schema;
              }
              case "string": {
                return getStringAttributeParser(name, required).parse({ ...schema, type: "string" })
                  .schema;
              }
              case "array": {
                return getArrayAttributeParser(name, required).parse({ ...schema, type: "array" })
                  .schema;
              }
              case "object": {
                return getObjectAttributeParser(name, required).parse({ ...schema, type: "object" })
                  .schema;
              }
            }
          }),
          type: "multi",
        })
      )
  );
}

// Schema

type AnySchema =
  | BooleanSchema
  | IntegerSchema
  | NullSchema
  | NumberSchema
  | StringSchema
  | ArraySchemaInput
  | ObjectSchemaInput
  | MultiSchemaInput;

type SchemaInput = AnySchema & SchemaProps;

const schemaPropsParser = getStrictParser<SchemaProps>()(
  z.object({
    $metadata: z.object({ uris: z.object({ jsonLdContext: z.string() }) }),
  })
);

function getAttributeParser(name: string, required: boolean) {
  return getStrictParser<AnySchema, Attribute>()(
    z.union([
      getBooleanAttributeParser(name, required),
      getIntegerAttributeParser(name, required),
      getNullAttributeParser(name, required),
      getNumberAttributeParser(name, required),
      getStringAttributeParser(name, required),
      getArrayAttributeParser(name, required),
      getObjectAttributeParser(name, required),
      getMultiSchemaParser(name, required),
    ])
  );
}

export const jsonSchemaParser = getStrictParser<SchemaInput, JsonSchema>()(
  schemaPropsParser.and(getAttributeParser("schema", false))
);

// JSON LD Type

function getIden3JsonLdTypeParser(schema: JsonSchema) {
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
        const schemaCredentialSubject =
          schema.type === "object" && schema.schema.properties
            ? schema.schema.properties.reduce(
                (acc: ObjectAttribute | undefined, curr: Attribute) =>
                  curr.type === "object" && curr.name === "credentialSubject" ? curr : acc,
                undefined
              )
            : undefined;

        if (!schemaCredentialSubject) {
          zodContext.addIssue({
            code: z.ZodIssueCode.custom,
            fatal: true,
            message: "Couldn't find attribute credentialSubject in the schema",
          });
          return z.NEVER;
        }

        const ldContextTypeParseResult = Object.entries(ldContext["@context"][0]).reduce(
          (acc: { success: false } | { success: true; value: JsonLdType[] }, [key, value]) => {
            const parsedValue = z
              .object({
                "@context": z.record(z.unknown()),
                "@id": z.string().url("Property @id of the type is not a valid URL"),
              })
              .safeParse(value);

            const ldContextTypePropsParseResult = parsedValue.success
              ? schemaCredentialSubject.schema.properties?.reduce(
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

function getSertoJsonLdTypeParser(schema: JsonSchema) {
  return getStrictParser<
    {
      "@context": Record<string, unknown>;
    },
    JsonLdType[]
  >()(
    z
      .object({
        "@context": z.record(z.unknown()),
      })
      .transform((ldContext, zodContext): JsonLdType[] => {
        const schemaCredentialSubject =
          schema.type === "object" && schema.schema.properties
            ? schema.schema.properties.reduce(
                (acc: ObjectAttribute | undefined, curr: Attribute) =>
                  curr.type === "object" && curr.name === "credentialSubject" ? curr : acc,
                undefined
              )
            : undefined;
        if (!schemaCredentialSubject) {
          zodContext.addIssue({
            code: z.ZodIssueCode.custom,
            fatal: true,
            message: "Couldn't find attribute credentialSubject in the schema",
          });

          return z.NEVER;
        }

        const { credentialSubject: ldContextCredentialSubject, "schema-id": schemaId } = z
          .object({
            credentialSubject: z.object({
              "@context": z.record(z.unknown()),
            }),
            "schema-id": z.string().url(),
          })
          .parse(ldContext["@context"]);

        const jsonLdTypeParseResult = Object.entries(ldContext["@context"]).reduce(
          (acc: { success: false } | { jsonLdType: JsonLdType; success: true }, [key, value]) => {
            const parsedValue = z.object({ "@id": z.literal("schema-id") }).safeParse(value);

            return !acc.success && parsedValue.success
              ? {
                  jsonLdType: { id: `${schemaId}${key}`, name: key },
                  success: true,
                }
              : acc;
          },
          { success: false }
        );

        const ldContextTypePropsParseResult = schemaCredentialSubject.schema.properties?.reduce(
          (acc: { success: true } | { error: string; success: false }, attribute) =>
            acc.success && attribute.name in ldContextCredentialSubject["@context"]
              ? acc
              : {
                  error: `Couldn't find Property "${attribute.name}" of the JSON schema in the context`,
                  success: false,
                },
          { success: true }
        ) || {
          error: "Couldn't find any properties in schema's credentialSubject",
          success: false,
        };

        if (jsonLdTypeParseResult.success && ldContextTypePropsParseResult.success) {
          return [jsonLdTypeParseResult.jsonLdType];
        } else {
          zodContext.addIssue({
            code: z.ZodIssueCode.custom,
            fatal: true,
            message: !ldContextTypePropsParseResult.success
              ? ldContextTypePropsParseResult.error
              : "Couldn't find any valid type in the JSON LD context of the schema",
          });

          return z.NEVER;
        }
      })
  );
}

export function getJsonLdTypeParser(schema: JsonSchema) {
  return getStrictParser<
    | {
        "@context": Record<string, unknown>;
      }
    | {
        "@context": [Record<string, unknown>];
      },
    JsonLdType[]
  >()(z.union([getIden3JsonLdTypeParser(schema), getSertoJsonLdTypeParser(schema)]));
}
