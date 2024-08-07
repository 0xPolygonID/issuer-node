import { message } from "antd";
import {
  PropsWithChildren,
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { useEnvContext } from "./Env";
import { getApiIssuers } from "src/adapters/api/issuers";
import { AppError } from "src/domain";
import { Identifier, Issuer } from "src/domain/identifier";
import { AsyncTask, isAsyncTaskDataAvailable } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { IDENTIFIER_LOCAL_STORAGE_KEY } from "src/utils/constants";

type IssuerState = {
  fetchIssuers: (signal: AbortSignal) => void;
  handleChange: (identifier: Identifier) => void;
  identifier: Identifier;
  issuersList: AsyncTask<Issuer[], AppError>;
};

const defaultIssuerState: IssuerState = {
  fetchIssuers: function (): void {},
  handleChange: function (): void {},
  identifier: null,
  issuersList: { status: "pending" },
};

const IssuerContext = createContext(defaultIssuerState);

export function IssuerProvider(props: PropsWithChildren) {
  const env = useEnvContext();
  const [messageAPI, messageContext] = message.useMessage();
  const [issuersList, setIssuersList] = useState<AsyncTask<Issuer[], AppError>>({
    status: "pending",
  });
  const [identifier, setIdentifier] = useState<Identifier>(null);

  const fetchIssuers = useCallback(
    async (signal: AbortSignal) => {
      setIssuersList((previousState) =>
        isAsyncTaskDataAvailable(previousState)
          ? { data: previousState.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await getApiIssuers({
        env,
        signal,
      });

      if (response.success) {
        const issuers = response.data.successful;
        const [firstIssuer] = issuers;
        const savedIdentifier = window.localStorage.getItem(IDENTIFIER_LOCAL_STORAGE_KEY);

        setIssuersList({ data: issuers, status: "successful" });

        if (issuers.length === 1 && firstIssuer) {
          setIdentifier(firstIssuer.identifier);
        } else if (
          issuers.length > 1 &&
          savedIdentifier &&
          issuers.some(({ identifier }) => identifier === savedIdentifier)
        ) {
          setIdentifier(savedIdentifier);
        }
      } else {
        if (!isAbortedError(response.error)) {
          setIssuersList({ error: response.error, status: "failed" });
          void messageAPI.error(response.error.message);
        }
      }
    },
    [env, messageAPI]
  );

  const handleChange = useCallback((identifier: Identifier) => {
    setIdentifier(identifier);
  }, []);

  useEffect(() => {
    if (identifier) {
      window.localStorage.setItem(IDENTIFIER_LOCAL_STORAGE_KEY, identifier);
    }
  }, [identifier]);

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchIssuers);

    return aborter;
  }, [fetchIssuers]);

  const value = useMemo(() => {
    return { fetchIssuers, handleChange, identifier, issuersList };
  }, [identifier, issuersList, handleChange, fetchIssuers]);

  return (
    <>
      {messageContext}
      {(issuersList.status === "successful" || issuersList.status === "reloading") && (
        <IssuerContext.Provider value={value} {...props} />
      )}
    </>
  );
}

export function useIssuerContext() {
  return useContext(IssuerContext);
}
