import dayjs from "dayjs";

export function formatDate(
  date: dayjs.Dayjs | Date,
  format: "date" | "date-time" | "time" = "date"
) {
  const template =
    format === "date" ? "YYYY-MM-DD" : format === "date-time" ? "YYYY-MM-DD HH:mm" : "HH:mm:ss";

  return dayjs(date).format(template);
}
