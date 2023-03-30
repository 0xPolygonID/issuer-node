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
  MultiAttribute,
  MultiSchema,
  NullAttribute,
  NullSchema,
  NumberAttribute,
  NumberSchema,
  ObjectAttribute,
  ObjectProps,
  Schema,
  SchemaProps,
  StringAttribute,
  StringProps,
  StringSchema,
} from "src/domain/schemas";
import { StrictSchema } from "src/utils/types";

const commonPropsParser = StrictSchema<CommonProps>()(
  z.object({
    description: z.string().optional(),
    title: z.string().optional(),
  })
);

const stringPropsParser = StrictSchema<StringProps>()(
  z.object({
    enum: z.string().array().min(1).optional(),
    format: z.string().optional(),
  })
);

const stringSchemaParser = StrictSchema<StringSchema>()(
  commonPropsParser.and(stringPropsParser).and(
    z.object({
      type: z.literal("string"),
    })
  )
);

const stringAttributeParser = (name: string, required: boolean) =>
  StrictSchema<StringSchema, StringAttribute>()(
    stringSchemaParser.transform(
      (schema): StringAttribute => ({
        name,
        required,
        schema,
        type: "string",
      })
    )
  );

const integerSchemaParser = StrictSchema<IntegerSchema>()(
  commonPropsParser.and(
    z.object({
      enum: z.number().int().array().min(1).optional(),
      type: z.literal("integer"),
    })
  )
);

const integerAttributeParser = (name: string, required: boolean) =>
  StrictSchema<IntegerSchema, IntegerAttribute>()(
    integerSchemaParser.transform(
      (schema): IntegerAttribute => ({
        name,
        required,
        schema,
        type: "integer",
      })
    )
  );

const numberSchemaParser = StrictSchema<NumberSchema>()(
  commonPropsParser.and(
    z.object({
      enum: z.number().array().min(1).optional(),
      type: z.literal("number"),
    })
  )
);

const numberAttributeParser = (name: string, required: boolean) =>
  StrictSchema<NumberSchema, NumberAttribute>()(
    numberSchemaParser.transform(
      (schema): NumberAttribute => ({
        name,
        required,
        schema,
        type: "number",
      })
    )
  );

const booleanPropsParser = StrictSchema<BooleanProps>()(
  z.object({
    enum: z.boolean().array().min(1).optional(),
  })
);

const booleanSchemaParser = StrictSchema<BooleanSchema>()(
  commonPropsParser.and(booleanPropsParser).and(
    z.object({
      type: z.literal("boolean"),
    })
  )
);

const booleanAttributeParser = (name: string, required: boolean) =>
  StrictSchema<BooleanSchema, BooleanAttribute>()(
    booleanSchemaParser.transform(
      (schema): BooleanAttribute => ({
        name,
        required,
        schema,
        type: "boolean",
      })
    )
  );

const nullSchemaParser = StrictSchema<NullSchema>()(
  commonPropsParser.and(
    z.object({
      type: z.literal("null"),
    })
  )
);

const nullAttributeParser = (name: string, required: boolean) =>
  StrictSchema<NullSchema, NullAttribute>()(
    nullSchemaParser.transform(
      (schema): NullAttribute => ({
        name,
        required,
        schema,
        type: "null",
      })
    )
  );

type ObjectPropsWithoutProperties = Omit<ObjectProps, "properties"> & {
  properties?: Record<string, unknown>;
};

type ObjectSchema = CommonProps & ObjectPropsWithoutProperties & { type: "object" };

const objectPropsParser = StrictSchema<ObjectPropsWithoutProperties>()(
  z.object({
    properties: z.record(z.unknown()).optional(),
    required: z.string().array().optional(),
  })
);

const objectSchemaParser = StrictSchema<ObjectSchema>()(
  commonPropsParser.and(objectPropsParser).and(
    z.object({
      type: z.literal("object"),
    })
  )
);

const objectAttributeParser = (name: string, required: boolean) =>
  StrictSchema<ObjectSchema, ObjectAttribute>()(
    objectSchemaParser.transform(
      (schema): ObjectAttribute => ({
        name,
        required,
        schema: {
          ...schema,
          properties: schema.properties
            ? Object.entries(schema.properties).map(([name, value]) =>
                attributeParser(
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

type ArrayPropsWithoutItems = Omit<ArrayProps, "items"> & {
  items?: unknown;
};

type ArraySchema = CommonProps & ArrayPropsWithoutItems & { type: "array" };

const arrayPropsParser = StrictSchema<ArrayPropsWithoutItems>()(
  z.object({
    items: z.unknown().optional(),
  })
);

const arraySchemaParser = StrictSchema<ArraySchema>()(
  commonPropsParser.and(arrayPropsParser).and(
    z.object({
      type: z.literal("array"),
    })
  )
);

const arrayAttributeParser = (name: string, required: boolean) =>
  StrictSchema<ArraySchema, ArrayAttribute>()(
    arraySchemaParser.transform(
      (schema): ArrayAttribute => ({
        name,
        required,
        schema: {
          ...schema,
          items: schema.items ? attributeParser("items", required).parse(schema.items) : undefined,
        },
        type: "array",
      })
    )
  );

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

const multiPropsParser = StrictSchema<MultiProps>()(
  stringPropsParser.and(booleanPropsParser).and(objectPropsParser).and(arrayPropsParser)
);

const multiSchemaTypeParser = StrictSchema<MultiSchemaType>()(
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

const unsafelyParseMultiSchemaToDomain = (
  multiSchema: MultiSchemaComposite,
  name: string,
  required: boolean
): MultiSchema => {
  return multiSchema.type.map((type): MultiSchema[number] => {
    switch (type) {
      case "string": {
        return stringAttributeParser(name, required).parse({ ...multiSchema, type: "string" })
          .schema;
      }
      case "integer": {
        return integerAttributeParser(name, required).parse({ ...multiSchema, type: "integer" })
          .schema;
      }
      case "number": {
        return numberAttributeParser(name, required).parse({ ...multiSchema, type: "number" })
          .schema;
      }
      case "boolean": {
        return booleanAttributeParser(name, required).parse({ ...multiSchema, type: "boolean" })
          .schema;
      }
      case "null": {
        return nullAttributeParser(name, required).parse({ ...multiSchema, type: "null" }).schema;
      }
      case "object": {
        return objectAttributeParser(name, required).parse({ ...multiSchema, type: "object" })
          .schema;
      }
      case "array": {
        return arrayAttributeParser(name, required).parse({ ...multiSchema, type: "array" }).schema;
      }
    }
  });
};

const multiSchemaParser = (name: string, required: boolean) =>
  StrictSchema<MultiSchemaComposite, MultiAttribute>()(
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

type AnySchema =
  | StringSchema
  | IntegerSchema
  | NumberSchema
  | BooleanSchema
  | NullSchema
  | ObjectSchema
  | ArraySchema
  | MultiSchemaComposite;

const attributeParser = (name: string, required: boolean) =>
  StrictSchema<AnySchema, Attribute>()(
    z.union([
      stringAttributeParser(name, required),
      integerAttributeParser(name, required),
      numberAttributeParser(name, required),
      booleanAttributeParser(name, required),
      nullAttributeParser(name, required),
      objectAttributeParser(name, required),
      arrayAttributeParser(name, required),
      multiSchemaParser(name, required),
    ])
  );

type SchemaComposite = AnySchema & SchemaProps;

const schemaPropsParser = StrictSchema<SchemaProps>()(
  z.object({
    $metadata: z.object({ uris: z.object({ jsonLdContext: z.string() }) }),
  })
);

const sertoJsonLdTypeParser = (schema: Schema) =>
  StrictSchema<
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

const iden3JsonLdTypeParser = (schema: Schema) =>
  StrictSchema<
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

export const schemaParser = StrictSchema<SchemaComposite, Schema>()(
  schemaPropsParser.and(attributeParser("schema", false))
);

export function jsonLdTypeParser(schema: Schema) {
  return StrictSchema<
    | {
        "@context": Record<string, unknown>;
      }
    | {
        "@context": [Record<string, unknown>];
      },
    JsonLdType[]
  >()(z.union([sertoJsonLdTypeParser(schema), iden3JsonLdTypeParser(schema)]));
}
