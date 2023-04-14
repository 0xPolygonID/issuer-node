import axios from "axios";
import z from "zod";

import { getStrictParser } from "src/adapters/parsers";
import { Json, JsonLiteral } from "src/domain";

const jsonLiteralParser = getStrictParser<JsonLiteral>()(
  z.union([z.string(), z.number(), z.boolean(), z.null()])
);

export const jsonParser: z.ZodType<Json> = getStrictParser<Json>()(
  z.lazy(() => z.union([jsonLiteralParser, z.array(jsonParser), z.record(jsonParser)]))
);

export async function downloadJsonFromUrl({ fileName, url }: { fileName: string; url: string }) {
  const json = await getJsonFromUrl({
    url: url,
  });
  const data = "data:text/json;charset=utf-8," + encodeURIComponent(JSON.stringify(json, null, 4));
  const a = document.createElement("a");

  a.setAttribute("href", data);
  a.setAttribute("download", fileName + ".json");
  document.body.appendChild(a); // required for Firefox
  a.click();
  a.remove();
}

export async function getJsonFromUrl({ signal, url }: { signal?: AbortSignal; url: string }) {
  const response = await axios({
    method: "GET",
    signal,
    url,
  });

  return jsonParser.parse(response.data);
}
