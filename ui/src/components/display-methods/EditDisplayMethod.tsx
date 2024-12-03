import { App, Card, Space } from "antd";
import { useCallback, useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";

import {
  UpsertDisplayMethod,
  getDisplayMethod,
  updateDisplayMethod,
} from "src/adapters/api/display-method";
import { DisplayMethodForm } from "src/components/display-methods/DisplayMethodForm";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { AppError, DisplayMethod } from "src/domain";
import { ROUTES } from "src/routes";
import {
  AsyncTask,
  hasAsyncTaskFailed,
  isAsyncTaskDataAvailable,
  isAsyncTaskStarting,
} from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { DISPLAY_METHOD_EDIT } from "src/utils/constants";

export function EditDisplayMethod() {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();
  const navigate = useNavigate();
  const { message } = App.useApp();
  const { displayMethodID } = useParams();

  const [displayMethod, setDisplayMethod] = useState<AsyncTask<DisplayMethod, AppError>>({
    status: "pending",
  });

  const fetchDisplayMethod = useCallback(
    async (signal?: AbortSignal) => {
      setDisplayMethod((previousDisplayMethod) =>
        isAsyncTaskDataAvailable(previousDisplayMethod)
          ? { data: previousDisplayMethod.data, status: "reloading" }
          : { status: "loading" }
      );

      if (displayMethodID) {
        const response = await getDisplayMethod({
          displayMethodID,
          env,
          identifier,
          signal,
        });

        if (response.success) {
          setDisplayMethod({
            data: response.data,
            status: "successful",
          });
        } else {
          if (!isAbortedError(response.error)) {
            setDisplayMethod({ error: response.error, status: "failed" });
          }
        }
      }
    },
    [env, identifier, displayMethodID]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchDisplayMethod);

    return aborter;
  }, [fetchDisplayMethod]);

  const handleSubmit = (formValues: UpsertDisplayMethod) => {
    if (displayMethodID) {
      return void updateDisplayMethod({
        env,
        id: displayMethodID,
        identifier,
        payload: { ...formValues, name: formValues.name.trim() },
      }).then((response) => {
        if (response.success) {
          void message.success("Display method edited successfully");
          navigate(ROUTES.displayMethods.path);
        } else {
          void message.error(response.error.message);
        }
      });
    }
  };

  if (!displayMethodID) {
    return <ErrorResult error="No display method provided." />;
  }

  return (
    <SiderLayoutContent
      description="Modify and update the settings of an existing display method"
      showBackButton
      showDivider
      title={DISPLAY_METHOD_EDIT}
    >
      {(() => {
        if (hasAsyncTaskFailed(displayMethod)) {
          return (
            <Card className="centered">
              <ErrorResult
                error={[
                  "An error occurred while downloading an display method from the API:",
                  displayMethod.error.message,
                ].join("\n")}
              />
            </Card>
          );
        } else if (isAsyncTaskStarting(displayMethod)) {
          return (
            <Card className="centered">
              <LoadingResult />
            </Card>
          );
        } else {
          return (
            <Card className="centered" title="Display method details">
              <Space direction="vertical" size="large">
                <DisplayMethodForm
                  initialValues={{
                    name: displayMethod.data.name,
                    url: displayMethod.data.url,
                  }}
                  onSubmit={handleSubmit}
                />
              </Space>
            </Card>
          );
        }
      })()}
    </SiderLayoutContent>
  );
}
