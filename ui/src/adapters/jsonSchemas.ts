import { RequestResponse, buildAppError } from "./api";
import { getJsonFromUrl } from "src/adapters/json";
import { getJsonLdTypeParser, jsonSchemaParser } from "src/adapters/parsers/jsonSchemas";
import { Json, JsonLdType, JsonSchema } from "src/domain";

export async function getJsonSchemaFromUrl({
  signal,
  url,
}: {
  signal?: AbortSignal;
  url: string;
}): Promise<RequestResponse<[JsonSchema, Json]>> {
  try {
    const jsonResponse = await getJsonFromUrl({
      signal,
      url,
    });
    if (!jsonResponse.isSuccessful) {
      return jsonResponse;
    } else {
      const json = jsonResponse.data;
      const jsonSchema = jsonSchemaParser.parse(json);
      return {
        data: [jsonSchema, json],
        isSuccessful: true,
      };
    }
  } catch (error) {
    return {
      error: buildAppError(error),
      isSuccessful: false,
    };
  }
}

export async function getSchemaJsonLdTypes({
  jsonSchema,
}: {
  jsonSchema: JsonSchema;
}): Promise<RequestResponse<[JsonLdType[], Json]>> {
  try {
    const jsonResponse = await getJsonFromUrl({
      url: jsonSchema.$metadata.uris.jsonLdContext,
    });
    if (!jsonResponse.isSuccessful) {
      return jsonResponse;
    } else {
      const json = jsonResponse.data;
      const jsonLdTypes = getJsonLdTypeParser(jsonSchema).parse(json);
      return {
        data: [jsonLdTypes, json],
        isSuccessful: true,
      };
    }
  } catch (error) {
    return {
      error: buildAppError(error),
      isSuccessful: false,
    };
  }
}
