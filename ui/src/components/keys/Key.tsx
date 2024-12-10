import { Card, Space, Typography } from "antd";
import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";

import { useIdentityContext } from "../../contexts/Identity";
import { getKey } from "src/adapters/api/keys";
import {} from "src/adapters/parsers/view";
import { Detail } from "src/components/shared/Detail";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { AppError, Key as KeyType } from "src/domain";
import { AsyncTask, hasAsyncTaskFailed, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { KEY_DETAILS } from "src/utils/constants";

export function Key() {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();

  const [key, setKey] = useState<AsyncTask<KeyType, AppError>>({
    status: "pending",
  });

  const { keyID } = useParams();

  const fetchKey = useCallback(
    async (signal: AbortSignal) => {
      if (keyID) {
        setKey({ status: "loading" });

        const response = await getKey({
          env,
          identifier,
          keyID,
          signal,
        });

        if (response.success) {
          setKey({ data: response.data, status: "successful" });
        } else {
          if (!isAbortedError(response.error)) {
            setKey({ error: response.error, status: "failed" });
          }
        }
      }
    },
    [env, keyID, identifier]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchKey);

    return aborter;
  }, [fetchKey]);

  if (!identifier) {
    return <ErrorResult error="No identifier provided." />;
  }

  return (
    <SiderLayoutContent
      description="View key details"
      showBackButton
      showDivider
      title={KEY_DETAILS}
    >
      {(() => {
        if (hasAsyncTaskFailed(key)) {
          return (
            <Card className="centered">
              <ErrorResult
                error={[
                  "An error occurred while downloading a key details from the API:",
                  key.error.message,
                ].join("\n")}
              />
            </Card>
          );
        } else if (isAsyncTaskStarting(key)) {
          return (
            <Card className="centered">
              <LoadingResult />
            </Card>
          );
        } else {
          return (
            <Card className="centered" title="Key details">
              <Card className="background-grey">
                <Space direction="vertical">
                  <Typography.Text type="secondary">KEY DETAILS</Typography.Text>

                  <Detail
                    copyable
                    copyableText={key.data.publicKey}
                    ellipsisPosition={5}
                    label="Public key"
                    text={key.data.publicKey}
                  />

                  <Detail label="Type" text={key.data.keyType} />
                  <Detail label="Auth core clam" text={`${key.data.isAuthCoreClaim}`} />
                </Space>
              </Card>
            </Card>
          );
        }
      })()}
    </SiderLayoutContent>
  );
}
