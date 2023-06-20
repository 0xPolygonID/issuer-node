import { Response } from "src/adapters";
import { getIPFSGatewayUrl } from "src/adapters/api/schemas";
import { getJsonFromUrl } from "src/adapters/json";
import { getJsonLdTypeParser, jsonSchemaParser } from "src/adapters/parsers/jsonSchemas";
import { Env, Json, JsonLdType, JsonSchema } from "src/domain";
import { buildAppError } from "src/utils/error";

export async function getJsonSchemaFromUrl({
  signal,
  url,
}: {
  signal?: AbortSignal;
  url: string;
}): Promise<Response<[JsonSchema, Json]>> {
  try {
    const jsonResponse = await getJsonFromUrl({
      signal,
      url,
    });
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
  const contextIPFSGatewayUrlResponse = getIPFSGatewayUrl(
    env,
    jsonSchema.jsonSchemaProps.$metadata.uris.jsonLdContext
  );
  if (contextIPFSGatewayUrlResponse.success) {
    try {
      const jsonResponse = await getJsonFromUrl({
        url: contextIPFSGatewayUrlResponse.data,
      });
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
  } else {
    return contextIPFSGatewayUrlResponse;
  }
}
