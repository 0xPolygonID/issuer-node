import { Response } from "src/adapters";
import { getJsonFromUrl } from "src/adapters/json";
import { buildAppError } from "src/adapters/parsers";
import { getJsonLdTypeParser, jsonSchemaParser } from "src/adapters/parsers/jsonSchemas";
import { Env, Json, JsonLdType, JsonSchema } from "src/domain";

export async function getJsonSchemaFromUrl({
  env,
  signal,
  url,
}: {
  env: Env;
  signal?: AbortSignal;
  url: string;
}): Promise<Response<[JsonSchema, Json]>> {
  try {
    const jsonResponse = await getJsonFromUrl({ env, signal, url });
    if (!jsonResponse.success) {
      return jsonResponse;
    } else {
      const jsonSchemaObject = jsonResponse.data;
      const jsonSchema = jsonSchemaParser.parse(jsonSchemaObject);
      return {
        data: [jsonSchema, jsonSchemaObject],
        success: true,
      };
    }
  } catch (error) {
    return {
      error: buildAppError(error),
      success: false,
    };
  }
}

export async function getSchemaJsonLdTypes({
  env,
  jsonSchema,
}: {
  env: Env;
  jsonSchema: JsonSchema;
}): Promise<Response<[JsonLdType[], Json]>> {
  const url = jsonSchema.jsonSchemaProps.$metadata.uris.jsonLdContext;
  try {
    const jsonResponse = await getJsonFromUrl({ env, url });
    if (!jsonResponse.success) {
      return jsonResponse;
    } else {
      const jsonLdContextObject = jsonResponse.data;
      const jsonLdTypes = getJsonLdTypeParser(jsonSchema).parse(jsonLdContextObject);
      return {
        data: [jsonLdTypes, jsonLdContextObject],
        success: true,
      };
    }
  } catch (error) {
    return {
      error: buildAppError(error),
      success: false,
    };
  }
}
