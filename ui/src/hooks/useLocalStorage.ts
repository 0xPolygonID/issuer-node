import ls from "localstorage-slim";
import { useEffect, useState } from "react";

export function useLocalStorage<T>(
  key: string,
  defaultValue: T
): [T, React.Dispatch<React.SetStateAction<T>>] {
  const [value, setValue] = useState(ls.get<T>(key) ?? defaultValue);

  useEffect(() => {
    ls.set(key, value);
  }, [key, value]);

  return [value, setValue];
}
