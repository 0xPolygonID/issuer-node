import { getJsonFromUrl } from "src/adapters/json";
import { getJsonLdTypeParser, jsonSchemaParser } from "src/adapters/parsers/jsonSchemas";
import { Json, JsonLdType, JsonSchema } from "src/domain";

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
