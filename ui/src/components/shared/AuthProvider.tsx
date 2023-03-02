import { PropsWithChildren, useCallback, useEffect, useMemo, useState } from "react";

import { Account, parseAccount } from "src/adapters/api/accounts";
import { issuersGet } from "src/adapters/api/issuers";
import { AuthContext, auth } from "src/contexts/auth";
import { Organization } from "src/domain";
import { APIError, HTTPStatusError } from "src/utils/adapters";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { AUTH_TOKEN_KEY } from "src/utils/constants";
import { AsyncTask } from "src/utils/types";

//TODO to be deleted after connected to localDB
export function AuthProvider(props: PropsWithChildren) {
  const storedAuthToken = localStorage.getItem(AUTH_TOKEN_KEY) || undefined;
  const storedAccount = storedAuthToken ? parseAccount(storedAuthToken) : undefined;

  const [authToken, setAuthToken] = useState<string | undefined>(storedAuthToken);
  const [account, setAccount] = useState<Account | undefined>(storedAccount);
  const [organization, setOrganization] = useState<AsyncTask<Organization, APIError>>({
    status: "pending",
  });

  const updateAuthToken = useCallback((newAuthToken: string) => {
    const newAccount = parseAccount(newAuthToken);

    setAuthToken(newAuthToken);
    setAccount(newAccount);
    localStorage.setItem(AUTH_TOKEN_KEY, newAuthToken);
  }, []);

  const removeAuthToken = useCallback(() => {
    setAuthToken(undefined);
    localStorage.removeItem(AUTH_TOKEN_KEY);
  }, []);

  const updateOrganization = useCallback((newOrganization: Organization) => {
    setOrganization({ data: newOrganization, status: "successful" });
  }, []);

  const getIssuers = useCallback(
    async (signal: AbortSignal) => {
      if (authToken && account?.organization) {
        const res = await issuersGet({
          id: account.organization,
          signal,
          token: authToken,
        });

        if (res.isSuccessful) {
          setOrganization({ data: res.data, status: "successful" });
        } else {
          if (!isAbortedError(res.error) && res.error.status !== HTTPStatusError.Unauthorized) {
            setOrganization({ error: res.error, status: "failed" });
          }
        }
      }
    },
    [account, authToken]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(getIssuers);

    return aborter;
  }, [getIssuers]);

  const value = useMemo<AuthContext>(() => {
    return {
      account,
      authToken,
      organization,
      removeAuthToken,
      updateAuthToken,
      updateOrganization,
    };
  }, [account, organization, authToken, updateAuthToken, removeAuthToken, updateOrganization]);

  return <auth.Provider value={value} {...props} />;
}
