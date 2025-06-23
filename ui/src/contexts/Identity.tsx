import { Space } from "antd";
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
import { notifyErrors } from "src/adapters/parsers";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
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

type IdentityState = {
  fetchIdentities: (signal?: AbortSignal) => void;
  getSelectedIdentity: () => Identity | undefined;
  identifier: string;
  identityList: AsyncTask<Identity[], AppError>;
  selectIdentity: (identifier: string) => void;
};

const defaultIdentityState: IdentityState = {
  fetchIdentities: () => void {},
  getSelectedIdentity: () => void {},
  identifier: "",
  identityList: { status: "pending" },
  selectIdentity: () => void {},
};

const IdentityContext = createContext(defaultIdentityState);

export function IdentityProvider(props: PropsWithChildren) {
  const env = useEnvContext();
  const navigate = useNavigate();
  const location = useLocation();
  const [identityList, setIdentityList] = useState<AsyncTask<Identity[], AppError>>({
    status: "pending",
  });
  const [identifier, setIdentifier] = useState("");
  const [searchParams, setSearchParams] = useSearchParams();

  const identifierParam = searchParams.get(IDENTIFIER_SEARCH_PARAM);

  const fetchIdentities = useCallback(
    async (signal?: AbortSignal) => {
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
          void notifyErrors(response.data.failed);
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
        }
      }
    },
    [env, identifierParam]
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
      identityList,
      selectIdentity,
    };
  }, [identifier, identityList, selectIdentity, fetchIdentities, getSelectedIdentity]);

  switch (identityList.status) {
    case "successful":
    case "reloading": {
      return <IdentityContext.Provider value={value} {...props} />;
    }
    case "failed": {
      return <ErrorResult error={identityList.error.message} />;
    }
    case "pending":
    case "loading": {
      return (
        <Space
          style={{
            alignItems: "center",
            display: "flex",
            height: "100vh",
            justifyContent: "center",
            width: "100vw",
          }}
        >
          <LoadingResult />
        </Space>
      );
    }
  }
}

export function useIdentityContext() {
  return useContext(IdentityContext);
}
