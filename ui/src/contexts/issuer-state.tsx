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

import { APIError } from "src/adapters/api";
import { getStatus } from "src/adapters/api/issuer-state";
import { useEnvContext } from "src/contexts/env";
import { AsyncTask } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";

interface State {
  notifyChange: () => Promise<void>;
  refreshStatus: () => Promise<void>;
  status: AsyncTask<boolean, APIError>;
}

const StateContext = createContext<State>({
  notifyChange: () => Promise.reject("The state context is not yet ready"),
  refreshStatus: () => Promise.reject("The state context is not yet ready"),
  status: { status: "pending" },
});

export function StateProvider(props: PropsWithChildren) {
  const env = useEnvContext();
  const [status, setStatus] = useState<AsyncTask<boolean, APIError>>({ status: "pending" });

  const refreshStatus = useCallback(
    async (signal?: AbortSignal) => {
      const response = await getStatus({ env, signal });

      if (response.isSuccessful) {
        setStatus({ data: response.data, status: "successful" });
      } else {
        if (!isAbortedError(response.error)) {
          void message.error(response.error.message);
        }
      }
    },
    [env]
  );

  const notifyChange = useCallback(() => {
    void message.success({
      content: (
        <Space align="start" direction="vertical" style={{ width: "auto" }}>
          <Typography.Text strong>Revocation requires issuer state to be published</Typography.Text>
          <Typography.Text type="secondary">
            Publish issuer state now or bulk publish with other actions.
          </Typography.Text>
        </Space>
      ),
    });

    return refreshStatus();
  }, [refreshStatus]);

  useEffect(() => {
    const { aborter } = makeRequestAbortable(refreshStatus);

    return aborter;
  }, [refreshStatus]);

  const value = useMemo(() => {
    return { notifyChange, refreshStatus, status };
  }, [notifyChange, refreshStatus, status]);

  return <StateContext.Provider value={value} {...props} />;
}

export function useStateContext() {
  return useContext(StateContext);
}
