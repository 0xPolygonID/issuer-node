import { getJsonFromUrl } from "src/adapters/json";
import { jsonLdTypeParser, schemaParser } from "src/adapters/parsers/schemas";
import { Json, JsonLdType, Schema } from "src/domain";

export function getSchemaFromUrl({ url }: { url: string }): Promise<[Schema, Json]> {
  return getJsonFromUrl({
    url,
  }).then((json) => [schemaParser.parse(json), json]);
}

export function getSchemaJsonLdTypes({
  schema,
}: {
  schema: Schema;
}): Promise<[JsonLdType[], Json]> {
  return getJsonFromUrl({
    url: schema.$metadata.uris.jsonLdContext,
  }).then((json) => [jsonLdTypeParser(schema).parse(json), json]);
}
