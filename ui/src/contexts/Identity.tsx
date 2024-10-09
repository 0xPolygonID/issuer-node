import { App } from "antd";
import {
  PropsWithChildren,
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { useLocation, useNavigate, useSearchParams } from "react-router-dom";
import { getIdentities, identifierParser } from "src/adapters/api/identities";
import { useEnvContext } from "src/contexts/Env";
import { AppError, Identifier, Identity } from "src/domain";
import { ROUTES } from "src/routes";
import { AsyncTask, isAsyncTaskDataAvailable } from "src/utils/async";
import {
  getStorageByKey,
  isAbortedError,
  makeRequestAbortable,
  setStorageByKey,
} from "src/utils/browser";
import {
  IDENTIFIER_LOCAL_STORAGE_KEY,
  IDENTIFIER_SEARCH_PARAM,
  ROOT_PATH,
} from "src/utils/constants";

type IdentityState = {
  fetchIdentities: (signal: AbortSignal) => void;
  handleChange: (identifier: Identifier) => void;
  identifier: Identifier;
  identitiesList: AsyncTask<Identity[], AppError>;
  identityDisplayName: string;
};

const defaultIdentityState: IdentityState = {
  fetchIdentities: () => void {},
  handleChange: () => void {},
  identifier: "",
  identitiesList: { status: "pending" },
  identityDisplayName: "",
};

const IdentityContext = createContext(defaultIdentityState);

export function IdentityProvider(props: PropsWithChildren) {
  const env = useEnvContext();
  const { message } = App.useApp();

  const navigate = useNavigate();
  const location = useLocation();
  const [identitiesList, setIdentitiesList] = useState<AsyncTask<Identity[], AppError>>({
    status: "pending",
  });
  const [identifier, setIdentifier] = useState<Identifier>("");
  const [searchParams, setSearchParams] = useSearchParams();

  const identity =
    (identitiesList.status === "successful" || identitiesList.status === "reloading") &&
    identitiesList.data.find((identity) => identity.identifier === identifier);
  const identityDisplayName = identity ? identity.displayName : "";

  const identifierParam = searchParams.get(IDENTIFIER_SEARCH_PARAM);

  const fetchIdentities = useCallback(
    async (signal: AbortSignal) => {
      setIdentitiesList((previousState) =>
        isAsyncTaskDataAvailable(previousState)
          ? { data: previousState.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await getIdentities({
        env,
        signal,
      });

      if (response.success) {
        const identities = response.data.successful;
        const savedIdentifier = getStorageByKey({
          defaultValue: "",
          key: IDENTIFIER_LOCAL_STORAGE_KEY,
          parser: identifierParser,
        });

        setIdentitiesList({ data: identities, status: "successful" });
        if (
          identifierParam &&
          identities.some(({ identifier }) => identifier === identifierParam)
        ) {
          setIdentifier(identifierParam);
        } else if (identities.some(({ identifier }) => identifier === savedIdentifier)) {
          setIdentifier(savedIdentifier);
        } else if (identities.length > 0 && identities[0]) {
          setIdentifier(identities[0].identifier);
        }
      } else {
        if (!isAbortedError(response.error)) {
          setIdentitiesList({ error: response.error, status: "failed" });
          void message.error(response.error.message);
        }
      }
    },
    [env, message, identifierParam]
  );

  const handleChange = useCallback(
    (identifier: Identifier) => {
      setIdentifier(identifier);
      navigate(ROUTES.schemas.path);
    },
    [navigate]
  );

  useEffect(() => {
    if (
      identifierParam &&
      identifier !== identifierParam &&
      isAsyncTaskDataAvailable(identitiesList) &&
      identitiesList.data.some((identity) => identity.identifier === identifierParam)
    ) {
      setIdentifier(identifierParam);
    } else if (identifier && identifier !== identifierParam && location.pathname !== ROOT_PATH) {
      setIdentifier(identifier);
      setSearchParams(
        (previousParams) => {
          const params = new URLSearchParams(previousParams);
          params.set(IDENTIFIER_SEARCH_PARAM, identifier);

          return params;
        },
        { replace: true }
      );
      setStorageByKey({ key: IDENTIFIER_LOCAL_STORAGE_KEY, value: identifier });
    }
  }, [identifier, identifierParam, identitiesList, location, setSearchParams]);

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchIdentities);

    return aborter;
  }, [fetchIdentities]);

  const value = useMemo(() => {
    return {
      fetchIdentities,
      handleChange,
      identifier,
      identitiesList,
      identityDisplayName,
    };
  }, [identifier, identityDisplayName, identitiesList, handleChange, fetchIdentities]);

  return (
    (identitiesList.status === "successful" || identitiesList.status === "reloading") && (
      <IdentityContext.Provider value={value} {...props} />
    )
  );
}

export function useIdentityContext() {
  return useContext(IdentityContext);
}
