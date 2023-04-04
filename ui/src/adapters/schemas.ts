import { getJsonFromUrl } from "src/adapters/json";
import { getJsonLdTypeParser, jsonSchemaParser } from "src/adapters/parsers/schemas";
import { Json, JsonLdType } from "src/domain";
import { JsonSchema } from "src/domain/jsonSchema";

export async function getSchemaFromUrl({ url }: { url: string }): Promise<[JsonSchema, Json]> {
  const json = await getJsonFromUrl({
    url,
  });

  return [jsonSchemaParser.parse(json), json];
}

export async function getSchemaJsonLdTypes({
  jsonSchema,
}: {
  jsonSchema: JsonSchema;
}): Promise<[JsonLdType[], Json]> {
  const json = await getJsonFromUrl({
    url: jsonSchema.$metadata.uris.jsonLdContext,
  });

  return [getJsonLdTypeParser(jsonSchema).parse(json), json];
}
