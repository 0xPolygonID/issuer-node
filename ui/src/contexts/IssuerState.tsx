import { Space, Typography, message } from "antd";
import {
  PropsWithChildren,
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";

import { getStatus } from "src/adapters/api/issuer-state";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { AppError } from "src/domain";
import { AsyncTask } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";

type EventType = "credential" | "revoke";

type IssuerState = {
  notifyChange: (event: EventType) => Promise<void>;
  refreshStatus: () => Promise<void>;
  status: AsyncTask<boolean, AppError>;
};

const CONTEXT_NOT_READY_MESSAGE = "The issuer state context is not yet ready";

const IssuerStateContext = createContext<IssuerState>({
  notifyChange: () => Promise.reject(CONTEXT_NOT_READY_MESSAGE),
  refreshStatus: () => Promise.reject(CONTEXT_NOT_READY_MESSAGE),
  status: { status: "pending" },
});

export function IssuerStateProvider(props: PropsWithChildren) {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();

  const [messageAPI, messageContext] = message.useMessage();

  const [status, setStatus] = useState<AsyncTask<boolean, AppError>>({ status: "pending" });

  const refreshStatus = useCallback(
    async (signal?: AbortSignal) => {
      if (identifier) {
        const response = await getStatus({ env, identifier, signal });

        if (response.success) {
          setStatus({ data: response.data.pendingActions, status: "successful" });
        } else {
          if (!isAbortedError(response.error)) {
            void messageAPI.error(response.error.message);
          }
        }
      }
    },
    [env, messageAPI, identifier]
  );

  const notifyChange = useCallback(
    (event: EventType) => {
      const eventTitle: Record<EventType, string> = {
        credential: "Credential",
        revoke: "Revocation",
      };

      void messageAPI.info({
        content: (
          <Space align="start" direction="vertical" style={{ width: "auto" }}>
            <Typography.Text strong>
              {`${eventTitle[event]}`} requires issuer state to be published
            </Typography.Text>
            <Typography.Text type="secondary">
              Publish issuer state now or bulk publish with other actions.
            </Typography.Text>
          </Space>
        ),
      });

      return refreshStatus();
    },
    [messageAPI, refreshStatus]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(refreshStatus);

    return aborter;
  }, [refreshStatus]);

  const value = useMemo(() => {
    return { notifyChange, refreshStatus, status };
  }, [notifyChange, refreshStatus, status]);

  return (
    <>
      {messageContext}

      <IssuerStateContext.Provider value={value} {...props} />
    </>
  );
}

export function useIssuerStateContext() {
  return useContext(IssuerStateContext);
}
