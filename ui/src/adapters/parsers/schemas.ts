import { z } from "zod";

import { JsonLdType } from "src/domain";
import {
  ArrayAttribute,
  ArrayProps,
  Attribute,
  BooleanAttribute,
  BooleanProps,
  BooleanSchema,
  CommonProps,
  IntegerAttribute,
  IntegerSchema,
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
} from "src/domain/jsonSchema";
import { getStrictParser } from "src/utils/types";

// Types

type AnySchema =
  | StringSchema
  | IntegerSchema
  | NumberSchema
  | BooleanSchema
  | NullSchema
  | ObjectSchema
  | ArraySchema
  | MultiSchemaComposite;

type ArrayPropsWithoutItems = Omit<ArrayProps, "items"> & {
  items?: unknown;
};

type ArraySchema = CommonProps & ArrayPropsWithoutItems & { type: "array" };

type MultiSchemaType = {
  type: (
    | StringSchema["type"]
    | IntegerSchema["type"]
    | NumberSchema["type"]
    | BooleanSchema["type"]
    | NullSchema["type"]
    | ObjectSchema["type"]
    | ArraySchema["type"]
  )[];
};

type MultiProps = StringProps &
  BooleanProps &
  ObjectPropsWithoutProperties &
  ArrayPropsWithoutItems;

type MultiSchemaComposite = CommonProps & MultiProps & MultiSchemaType;

type ObjectPropsWithoutProperties = Omit<ObjectProps, "properties"> & {
  properties?: Record<string, unknown>;
};

type ObjectSchema = CommonProps & ObjectPropsWithoutProperties & { type: "object" };

type SchemaComposite = AnySchema & SchemaProps;

// Parsers

const commonPropsParser = getStrictParser<CommonProps>()(
  z.object({
    description: z.string().optional(),
    title: z.string().optional(),
  })
);

const arrayPropsParser = getStrictParser<ArrayPropsWithoutItems>()(
  z.object({
    items: z.unknown().optional(),
  })
);

const booleanPropsParser = getStrictParser<BooleanProps>()(
  z.object({
    enum: z.boolean().array().min(1).optional(),
  })
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

const objectPropsParser = getStrictParser<ObjectPropsWithoutProperties>()(
  z.object({
    properties: z.record(z.unknown()).optional(),
    required: z.string().array().optional(),
  })
);

const schemaPropsParser = getStrictParser<SchemaProps>()(
  z.object({
    $metadata: z.object({ uris: z.object({ jsonLdContext: z.string() }) }),
  })
);

const stringPropsParser = getStrictParser<StringProps>()(
  z.object({
    enum: z.string().array().min(1).optional(),
    format: z.string().optional(),
  })
);

// Composite parsers

const arraySchemaParser = getStrictParser<ArraySchema>()(
  commonPropsParser.and(arrayPropsParser).and(
    z.object({
      type: z.literal("array"),
    })
  )
);

const booleanSchemaParser = getStrictParser<BooleanSchema>()(
  commonPropsParser.and(booleanPropsParser).and(
    z.object({
      type: z.literal("boolean"),
    })
  )
);

const integerSchemaParser = getStrictParser<IntegerSchema>()(
  commonPropsParser.and(
    z.object({
      enum: z.number().int().array().min(1).optional(),
      type: z.literal("integer"),
    })
  )
);

const multiPropsParser = getStrictParser<MultiProps>()(
  stringPropsParser.and(booleanPropsParser).and(objectPropsParser).and(arrayPropsParser)
);

const nullSchemaParser = getStrictParser<NullSchema>()(
  commonPropsParser.and(
    z.object({
      type: z.literal("null"),
    })
  )
);

const numberSchemaParser = getStrictParser<NumberSchema>()(
  commonPropsParser.and(
    z.object({
      enum: z.number().array().min(1).optional(),
      type: z.literal("number"),
    })
  )
);

const objectSchemaParser = getStrictParser<ObjectSchema>()(
  commonPropsParser.and(objectPropsParser).and(
    z.object({
      type: z.literal("object"),
    })
  )
);

const stringSchemaParser = getStrictParser<StringSchema>()(
  commonPropsParser.and(stringPropsParser).and(
    z.object({
      type: z.literal("string"),
    })
  )
);

// Getters

const getArrayAttributeParser = (name: string, required: boolean) =>
  getStrictParser<ArraySchema, ArrayAttribute>()(
    arraySchemaParser.transform(
      (schema): ArrayAttribute => ({
        name,
        required,
        schema: {
          ...schema,
          items: schema.items
            ? getAttributeParser("items", required).parse(schema.items)
            : undefined,
        },
        type: "array",
      })
    )
  );

function getAttributeParser(name: string, required: boolean) {
  return getStrictParser<AnySchema, Attribute>()(
    z.union([
      getArrayAttributeParser(name, required),
      getBooleanAttributeParser(name, required),
      getIntegerAttributeParser(name, required),
      getMultiSchemaParser(name, required),
      getNumberAttributeParser(name, required),
      getNullAttributeParser(name, required),
      getObjectAttributeParser(name, required),
      getStringAttributeParser(name, required),
    ])
  );
}

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

function getMultiSchemaParser(name: string, required: boolean) {
  return getStrictParser<MultiSchemaComposite, MultiAttribute>()(
    commonPropsParser
      .and(multiPropsParser)
      .and(multiSchemaTypeParser)
      .transform(
        (schema): MultiAttribute => ({
          name,
          required,
          schemas: unsafelyParseMultiSchemaToDomain(schema, name, required),
          type: "multi",
        })
      )
  );
}

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

function getObjectAttributeParser(name: string, required: boolean) {
  return getStrictParser<ObjectSchema, ObjectAttribute>()(
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

// Helpers

const unsafelyParseMultiSchemaToDomain = (
  multiSchema: MultiSchemaComposite,
  name: string,
  required: boolean
): MultiSchema => {
  return multiSchema.type.map((type): MultiSchema[number] => {
    switch (type) {
      case "array": {
        return getArrayAttributeParser(name, required).parse({ ...multiSchema, type: "array" })
          .schema;
      }
      case "boolean": {
        return getBooleanAttributeParser(name, required).parse({ ...multiSchema, type: "boolean" })
          .schema;
      }
      case "integer": {
        return getIntegerAttributeParser(name, required).parse({ ...multiSchema, type: "integer" })
          .schema;
      }
      case "number": {
        return getNumberAttributeParser(name, required).parse({ ...multiSchema, type: "number" })
          .schema;
      }
      case "null": {
        return getNullAttributeParser(name, required).parse({ ...multiSchema, type: "null" })
          .schema;
      }
      case "object": {
        return getObjectAttributeParser(name, required).parse({ ...multiSchema, type: "object" })
          .schema;
      }
      case "string": {
        return getStringAttributeParser(name, required).parse({ ...multiSchema, type: "string" })
          .schema;
      }
    }
  });
};

// Exports

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

export const jsonSchemaParser = getStrictParser<SchemaComposite, JsonSchema>()(
  schemaPropsParser.and(getAttributeParser("schema", false))
);
