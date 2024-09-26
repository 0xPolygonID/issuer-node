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
import { getIssuers, identifierParser } from "src/adapters/api/issuers";
import { useEnvContext } from "src/contexts/Env";
import { AppError, Issuer, IssuerIdentifier } from "src/domain";
import { AsyncTask, isAsyncTaskDataAvailable } from "src/utils/async";
import {
  getStorageByKey,
  isAbortedError,
  makeRequestAbortable,
  setStorageByKey,
} from "src/utils/browser";
import { IDENTIFIER_LOCAL_STORAGE_KEY } from "src/utils/constants";

type IssuerState = {
  fetchIssuers: (signal: AbortSignal) => void;
  handleChange: (identifier: IssuerIdentifier) => void;
  issuerDisplayName: string;
  issuerIdentifier: IssuerIdentifier;
  issuersList: AsyncTask<Issuer[], AppError>;
};

const defaultIssuerState: IssuerState = {
  fetchIssuers: () => void {},
  handleChange: () => void {},
  issuerDisplayName: "",
  issuerIdentifier: "",
  issuersList: { status: "pending" },
};

const IssuerContext = createContext(defaultIssuerState);

export function IssuerProvider(props: PropsWithChildren) {
  const env = useEnvContext();
  const [messageAPI, messageContext] = message.useMessage();
  const [issuersList, setIssuersList] = useState<AsyncTask<Issuer[], AppError>>({
    status: "pending",
  });
  const [issuerIdentifier, setIssuerIdentifier] = useState<IssuerIdentifier>("");
  const issuer =
    issuersList.status === "successful" &&
    issuersList.data.find(({ identifier }) => identifier === issuerIdentifier);
  const issuerDisplayName = issuer ? issuer.displayName : "";

  const fetchIssuers = useCallback(
    async (signal: AbortSignal) => {
      setIssuersList((previousState) =>
        isAsyncTaskDataAvailable(previousState)
          ? { data: previousState.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await getIssuers({
        env,
        signal,
      });

      if (response.success) {
        const issuers = response.data.successful;
        const savedIdentifier = getStorageByKey({
          defaultValue: "",
          key: IDENTIFIER_LOCAL_STORAGE_KEY,
          parser: identifierParser,
        });

        setIssuersList({ data: issuers, status: "successful" });
        if (issuers.some(({ identifier }) => identifier === savedIdentifier)) {
          setIssuerIdentifier(savedIdentifier);
        } else if (issuers.length > 0 && issuers[0]) {
          setIssuerIdentifier(issuers[0].identifier);
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

  const handleChange = useCallback((identifier: IssuerIdentifier) => {
    setIssuerIdentifier(identifier);
  }, []);

  useEffect(() => {
    if (issuerIdentifier) {
      setStorageByKey({ key: IDENTIFIER_LOCAL_STORAGE_KEY, value: issuerIdentifier });
    }
  }, [issuerIdentifier]);

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchIssuers);

    return aborter;
  }, [fetchIssuers]);

  const value = useMemo(() => {
    return {
      fetchIssuers,
      handleChange,
      issuerDisplayName,
      issuerIdentifier,
      issuersList,
    };
  }, [issuerIdentifier, issuerDisplayName, issuersList, handleChange, fetchIssuers]);

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
