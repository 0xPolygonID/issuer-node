import { Attribute, JsonSchema, ObjectAttribute } from "src/domain";

export function extractCredentialSubjectAttribute(
  jsonSchema: JsonSchema
): ObjectAttribute | undefined {
  return jsonSchema.type === "object"
    ? jsonSchema.schema.attributes
        ?.filter((child): child is ObjectAttribute => child.type === "object")
        .find((child) => child.name === "credentialSubject")
    : undefined;
}

export function extractCredentialSubjectAttributeWithoutId(
  jsonSchema: JsonSchema
): ObjectAttribute | undefined {
  const credentialSubjectAttribute = extractCredentialSubjectAttribute(jsonSchema);
  return (
    credentialSubjectAttribute && {
      ...credentialSubjectAttribute,
      schema: {
        ...credentialSubjectAttribute.schema,
        attributes: credentialSubjectAttribute.schema.attributes?.filter(
          (attribute) => attribute.name !== "id"
        ),
        required:
          credentialSubjectAttribute.schema.required &&
          credentialSubjectAttribute.schema.required.filter((name) => name !== "id"),
      },
    }
  );
}

export function makeAttributeOptional(attribute: Attribute): Attribute {
  const { name, type } = attribute;
  switch (type) {
    case "boolean": {
      return {
        name,
        required: false,
        schema: attribute.schema,
        type,
      };
    }
    case "integer": {
      return { name, required: false, schema: attribute.schema, type };
    }
    case "null": {
      return { name, required: false, schema: attribute.schema, type };
    }
    case "number": {
      return { name, required: false, schema: attribute.schema, type };
    }
    case "string": {
      return { name, required: false, schema: attribute.schema, type };
    }
    case "array": {
      return {
        name,
        required: false,
        schema: {
          ...attribute.schema,
          items: attribute.schema.items && makeAttributeOptional(attribute.schema.items),
        },
        type,
      };
    }
    case "object": {
      return {
        name,
        required: false,
        schema: {
          ...attribute.schema,
          attributes: attribute.schema.attributes?.map(makeAttributeOptional),
        },
        type,
      };
    }
    case "multi": {
      return {
        name,
        required: false,
        schemas: attribute.schemas.map((schema) => {
          switch (schema.type) {
            case "object": {
              return {
                ...schema,
                attributes: schema.attributes?.map(makeAttributeOptional),
              };
            }
            case "array": {
              return {
                ...schema,
                items: schema.items && makeAttributeOptional(schema.items),
              };
            }
            default: {
              return schema;
            }
          }
        }),
        type,
      };
    }
  }
}
