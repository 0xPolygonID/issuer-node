import axios from "axios";
import z from "zod";

import { jsonLdTypeParser, schemaParser } from "src/adapters/parsers/schemas";
import { Json, JsonLdType, JsonLiteral, Schema } from "src/domain";
import { StrictSchema } from "src/utils/types";

export function processZodError<T>(error: z.ZodError<T>, init: string[] = []): string[] {
  return error.errors.reduce((mainAcc, issue): string[] => {
    switch (issue.code) {
      case "invalid_union": {
        return [
          ...mainAcc,
          ...issue.unionErrors.reduce(
            (innerAcc: string[], current: z.ZodError<T>): string[] => [
              ...innerAcc,
              ...processZodError(current, mainAcc),
            ],
            []
          ),
        ];
      }
      default: {
        const errorMsg = issue.path.length
          ? `${issue.message} at ${issue.path.join(".")}`
          : issue.message;
        return [...mainAcc, errorMsg];
      }
    }
  }, init);
}

const jsonLiteralParser = StrictSchema<JsonLiteral>()(
  z.union([z.string(), z.number(), z.boolean(), z.null()])
);
const jsonParser: z.ZodType<Json> = StrictSchema<Json>()(
  z.lazy(() => z.union([jsonLiteralParser, z.array(jsonParser), z.record(jsonParser)]))
);

function getJsonFromUrl({ url }: { url: string }): Promise<Json> {
  return axios({
    method: "GET",
    url: url,
  }).then((response) => {
    return jsonParser.parse(response.data);
  });
}

export function getSchemaFromUrl({ url }: { url: string }): Promise<[Schema, Json]> {
  return getJsonFromUrl({
    url,
  }).then((json) => [schemaParser.parse(json), json]);
}

export function getJsonLdTypesFromUrl({
  schema,
  url,
}: {
  schema: Schema;
  url: string;
}): Promise<[JsonLdType[], Json]> {
  return getJsonFromUrl({
    url,
  }).then((json) => [jsonLdTypeParser(schema).parse(json), json]);
}

export function downloadJsonFromUrl({
  fileName,
  url,
}: {
  fileName: string;
  url: string;
}): Promise<void> {
  return getJsonFromUrl({
    url: url,
  }).then((json) => {
    const data =
      "data:text/json;charset=utf-8," + encodeURIComponent(JSON.stringify(json, null, 4));
    const a = document.createElement("a");

    a.setAttribute("href", data);
    a.setAttribute("download", fileName + ".json");
    document.body.appendChild(a); // required for firefox
    a.click();
    a.remove();
  });
}
