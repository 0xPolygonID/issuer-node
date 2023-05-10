import { JsonSchema, ObjectAttribute } from "src/domain";

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
