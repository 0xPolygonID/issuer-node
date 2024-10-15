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
import { AppError, Identity } from "src/domain";
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
import { buildAppError } from "src/utils/error";

type IdentityState = {
  fetchIdentities: (signal: AbortSignal) => void;
  getSelectedIdentity: () => Identity | undefined;
  identifier: string;
  identityDisplayName: string;
  identityList: AsyncTask<Identity[], AppError>;
  selectIdentity: (identifier: string) => void;
};

const defaultIdentityState: IdentityState = {
  fetchIdentities: () => void {},
  getSelectedIdentity: () => void {},
  identifier: "",
  identityDisplayName: "",
  identityList: { status: "pending" },
  selectIdentity: () => void {},
};

const IdentityContext = createContext(defaultIdentityState);

export function IdentityProvider(props: PropsWithChildren) {
  const env = useEnvContext();
  const { message } = App.useApp();
  const navigate = useNavigate();
  const location = useLocation();
  const [identityList, setIdentityList] = useState<AsyncTask<Identity[], AppError>>({
    status: "pending",
  });
  const [identifier, setIdentifier] = useState("");
  const [searchParams, setSearchParams] = useSearchParams();

  const identity =
    (identityList.status === "successful" || identityList.status === "reloading") &&
    identityList.data.find((identity) => identity.identifier === identifier);
  const identityDisplayName = identity && identity.displayName ? identity.displayName : "";

  const identifierParam = searchParams.get(IDENTIFIER_SEARCH_PARAM);

  const fetchIdentities = useCallback(
    async (signal: AbortSignal) => {
      setIdentityList((previousState) =>
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

        if (response.data.failed.length) {
          void message.error(
            response.data.failed.map((error) => buildAppError(error).message).join("\n")
          );
        }

        const savedIdentifier = getStorageByKey({
          defaultValue: "",
          key: IDENTIFIER_LOCAL_STORAGE_KEY,
          parser: identifierParser,
        });

        setIdentityList({ data: identities, status: "successful" });
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
          setIdentityList({ error: response.error, status: "failed" });
          void message.error(response.error.message);
        }
      }
    },
    [env, message, identifierParam]
  );

  const selectIdentity = useCallback(
    (identifier: string) => {
      setIdentifier(identifier);
      navigate(ROUTES.schemas.path);
    },
    [navigate]
  );

  const getSelectedIdentity = useCallback(() => {
    return isAsyncTaskDataAvailable(identityList)
      ? identityList.data.find((identity) => identity.identifier === identifier)
      : undefined;
  }, [identifier, identityList]);

  useEffect(() => {
    if (
      identifierParam &&
      identifier !== identifierParam &&
      isAsyncTaskDataAvailable(identityList) &&
      identityList.data.some((identity) => identity.identifier === identifierParam)
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
  }, [identifier, identifierParam, identityList, location, setSearchParams]);

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchIdentities);

    return aborter;
  }, [fetchIdentities]);

  const value = useMemo(() => {
    return {
      fetchIdentities,
      getSelectedIdentity,
      identifier,
      identityDisplayName,
      identityList,
      selectIdentity,
    };
  }, [
    identifier,
    identityDisplayName,
    identityList,
    selectIdentity,
    fetchIdentities,
    getSelectedIdentity,
  ]);

  return (
    (identityList.status === "successful" || identityList.status === "reloading") && (
      <IdentityContext.Provider value={value} {...props} />
    )
  );
}

export function useIdentityContext() {
  return useContext(IdentityContext);
}
