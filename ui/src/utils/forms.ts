import dayjs from "dayjs";
import isAlphanumeric from "validator/lib/isAlphanumeric";
import isEmail from "validator/lib/isEmail";

const createFormValidator =
  (validator: (input: string) => boolean) => (_: unknown, value: unknown) => {
    if (typeof value !== "string") {
      return Promise.reject("The type of the input should be a string");
    } else if (value === "" || validator(value)) {
      return Promise.resolve();
    } else {
      return Promise.reject("Validation failed");
    }
  };

export const alphanumericValidator = createFormValidator(isAlphanumeric);

export const emailValidator = createFormValidator(isEmail);

export function formatDate(date: dayjs.Dayjs | Date, showTime?: boolean): string {
  const format = showTime ? "YYYY-MM-DD HH:mm" : "YYYY-MM-DD";
  return dayjs(date).format(format);
}
