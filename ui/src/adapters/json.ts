import axios from "axios";
import z from "zod";

import { Json, JsonLiteral } from "src/domain";
import { StrictSchema } from "src/utils/types";

const jsonLiteralParser = StrictSchema<JsonLiteral>()(
  z.union([z.string(), z.number(), z.boolean(), z.null()])
);

const jsonParser: z.ZodType<Json> = StrictSchema<Json>()(
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
  document.body.appendChild(a); // required for firefox
  a.click();
  a.remove();
}

export async function getJsonFromUrl({ url }: { url: string }) {
  const response = await axios({
    method: "GET",
    url: url,
  });

  return jsonParser.parse(response.data);
}
