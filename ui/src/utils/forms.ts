import dayjs from "dayjs";

export function formatDate(date: dayjs.Dayjs | Date, showTime?: boolean): string {
  const format = showTime ? "YYYY-MM-DD HH:mm" : "YYYY-MM-DD";
  return dayjs(date).format(format);
}
