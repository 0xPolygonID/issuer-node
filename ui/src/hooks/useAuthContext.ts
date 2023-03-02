import { useContext } from "react";
import { auth } from "src/contexts/auth";

export function useAuthContext() {
  return useContext(auth);
}
