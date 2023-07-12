import axios from "axios";
import z from "zod";

import { Response, buildErrorResponse, buildSuccessResponse } from "src/adapters";
import { processUrl } from "src/adapters/api/schemas";
import { getStrictParser } from "src/adapters/parsers";
import { Env, Json, JsonLiteral } from "src/domain";

const jsonLiteralParser = getStrictParser<JsonLiteral>()(
  z.union([z.string(), z.number(), z.boolean(), z.null()])
);

export const jsonParser: z.ZodType<Json> = getStrictParser<Json>()(
  z.lazy(() => z.union([jsonLiteralParser, z.array(jsonParser), z.record(jsonParser)]))
);

export async function downloadJsonFromUrl({
  env,
  fileName,
  url,
}: {
  env: Env;
  fileName: string;
  url: string;
}) {
  const json = await getJsonFromUrl({ env, url });
  const data = "data:text/json;charset=utf-8," + encodeURIComponent(JSON.stringify(json, null, 4));
  const a = document.createElement("a");

  a.setAttribute("href", data);
  a.setAttribute("download", fileName + ".json");
  document.body.appendChild(a); // required for Firefox
  a.click();
  a.remove();
}

export async function getJsonFromUrl({
  env,
  signal,
  url,
}: {
  env: Env;
  signal?: AbortSignal;
  url: string;
}): Promise<Response<Json>> {
  const processedUrl = processUrl(url, env);
  if (!processedUrl.success) {
    throw new Error(processedUrl.error.message);
  }

  try {
    const response = await axios({
      method: "GET",
      signal,
      url: processedUrl.data,
    });
    return buildSuccessResponse(jsonParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}
