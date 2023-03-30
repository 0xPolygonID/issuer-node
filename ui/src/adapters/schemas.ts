import { getJsonFromUrl } from "src/adapters/json";
import { jsonLdTypeParser, schemaParser } from "src/adapters/parsers/schemas";
import { Json, JsonLdType, Schema } from "src/domain";

export async function getSchemaFromUrl({ url }: { url: string }): Promise<[Schema, Json]> {
  const json = await getJsonFromUrl({
    url,
  });

  return [schemaParser.parse(json), json];
}

export async function getSchemaJsonLdTypes({
  schema,
}: {
  schema: Schema;
}): Promise<[JsonLdType[], Json]> {
  const json = await getJsonFromUrl({
    url: schema.$metadata.uris.jsonLdContext,
  });

  return [jsonLdTypeParser(schema).parse(json), json];
}
