import dayjs from "dayjs";
import { Identifier } from "src/domain";

export function formatDate(
  date: dayjs.Dayjs | Date,
  format: "date" | "date-time" | "time" = "date-time"
) {
  const template =
    format === "date" ? "YYYY-MM-DD" : format === "date-time" ? "YYYY-MM-DD HH:mm" : "HH:mm:ss";

  return dayjs(date).format(template);
}

export function formatIdentifier(identifier: Identifier): string {
  const parts = identifier.split(":");
  const id = parts.at(-1);
  const shortId = `${id?.slice(0, 5)}...${id?.slice(-4)}`;
  return parts.toSpliced(-1, 1, shortId).join(":");
}
