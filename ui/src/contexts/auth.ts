import { createContext } from "react";

import { Account } from "src/adapters/api/accounts";
import { Organization } from "src/domain";
import { APIError } from "src/utils/adapters";
import { AUTH_CONTEXT_NOT_READY_MESSAGE } from "src/utils/constants";
import { AsyncTask } from "src/utils/types";

export interface AuthContext {
  account?: Account;
  authToken?: string;
  organization: AsyncTask<Organization, APIError>;
  removeAuthToken: () => void;
  updateAuthToken: (authToken: string) => void;
  updateOrganization: (organization: Organization) => void;
}

export const auth = createContext<AuthContext>({
  organization: {
    status: "pending",
  },
  removeAuthToken: () => {
    throw new Error(AUTH_CONTEXT_NOT_READY_MESSAGE);
  },
  updateAuthToken: () => {
    throw new Error(AUTH_CONTEXT_NOT_READY_MESSAGE);
  },
  updateOrganization: () => {
    throw new Error(AUTH_CONTEXT_NOT_READY_MESSAGE);
  },
});
