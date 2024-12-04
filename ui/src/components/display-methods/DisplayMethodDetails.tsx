import { Button, Card, Divider, Flex, Space, Typography } from "antd";
import { useCallback, useEffect, useState } from "react";
import { generatePath, useNavigate, useParams } from "react-router-dom";

import { DISPLAY_METHOD_DETAILS, DISPLAY_METHOD_EDIT } from "../../utils/constants";
import { getDisplayMethod, getDisplayMethodMetadata } from "src/adapters/api/display-method";
import EditIcon from "src/assets/icons/edit-02.svg?react";
import { DisplayMethodCard } from "src/components/display-methods/DisplayMethodCard";
import { DisplayMethodErrorResult } from "src/components/display-methods/DisplayMethodErrorResult";
import { Detail } from "src/components/shared/Detail";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { AppError, DisplayMethod, DisplayMethodMetadata } from "src/domain";
import { ROUTES } from "src/routes";
import {
  AsyncTask,
  hasAsyncTaskFailed,
  isAsyncTaskDataAvailable,
  isAsyncTaskStarting,
} from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";

function Details({ name, type, url }: Omit<DisplayMethod, "id">) {
  return (
    <Space direction="vertical">
      <Typography.Text type="secondary">DISPLAY METHOD DETAILS</Typography.Text>
      <Detail label="Name" text={name} />
      <Detail copyable href={url} label="URL" text={url} />
      <Detail label="Type" text={type} />
    </Space>
  );
}

export function DisplayMethodDetails() {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();
  const { displayMethodID } = useParams();
  const navigate = useNavigate();

  const [displayMethod, setDisplayMethod] = useState<AsyncTask<DisplayMethod, AppError>>({
    status: "pending",
  });

  const [displayMethodMetadata, setDisplayMethodMetadata] = useState<
    AsyncTask<DisplayMethodMetadata, AppError>
  >({
    status: "pending",
  });

  const fetchDisplayMethodMetadata = useCallback(
    (url: string, signal: AbortSignal) => {
      setDisplayMethodMetadata({ status: "loading" });
      void getDisplayMethodMetadata({
        env,
        signal,
        url,
      }).then((response) => {
        if (response.success) {
          setDisplayMethodMetadata({ data: response.data, status: "successful" });
        } else {
          if (!isAbortedError(response.error)) {
            setDisplayMethodMetadata({ error: response.error, status: "failed" });
          }
        }
      });
    },
    [env]
  );

  const fetchDisplayMethod = useCallback(
    async (signal: AbortSignal) => {
      setDisplayMethod((previousDisplayMethod) =>
        isAsyncTaskDataAvailable(previousDisplayMethod)
          ? { data: previousDisplayMethod.data, status: "reloading" }
          : { status: "loading" }
      );

      if (!displayMethodID) {
        return;
      }

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
        fetchDisplayMethodMetadata(response.data.url, signal);
      } else {
        if (!isAbortedError(response.error)) {
          setDisplayMethod({ error: response.error, status: "failed" });
        }
      }
    },
    [env, identifier, displayMethodID, fetchDisplayMethodMetadata]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchDisplayMethod);

    return aborter;
  }, [fetchDisplayMethod]);

  if (!displayMethodID) {
    return <ErrorResult error="No display method provided." />;
  }

  return (
    <SiderLayoutContent
      description="View display method details"
      extra={
        <Button
          icon={<EditIcon />}
          onClick={() => navigate(generatePath(ROUTES.editDisplayMethod.path, { displayMethodID }))}
          type="primary"
        >
          {DISPLAY_METHOD_EDIT}
        </Button>
      }
      showBackButton
      showDivider
      title={DISPLAY_METHOD_DETAILS}
    >
      {(() => {
        if (hasAsyncTaskFailed(displayMethod)) {
          return (
            <Card className="centered">
              <ErrorResult
                error={[
                  "An error occurred while downloading a display method from the API:",
                  displayMethod.error.message,
                ].join("\n")}
              />
            </Card>
          );
        } else if (hasAsyncTaskFailed(displayMethodMetadata)) {
          return (
            <Card className="centered">
              {isAsyncTaskDataAvailable(displayMethod) && (
                <>
                  <Details {...displayMethod.data} />
                  <Divider />
                </>
              )}
              {displayMethodMetadata.error.type === "parse-error" ? (
                <DisplayMethodErrorResult
                  labelRetry={DISPLAY_METHOD_EDIT}
                  message={displayMethodMetadata.error.message}
                  onRetry={() =>
                    navigate(generatePath(ROUTES.editDisplayMethod.path, { displayMethodID }))
                  }
                />
              ) : (
                <ErrorResult
                  error={[
                    "An error occurred while downloading a display method from the API:",
                    displayMethodMetadata.error.message,
                  ].join("\n")}
                />
              )}
            </Card>
          );
        } else if (
          isAsyncTaskStarting(displayMethod) ||
          isAsyncTaskStarting(displayMethodMetadata)
        ) {
          return (
            <Card className="centered">
              <LoadingResult />
            </Card>
          );
        } else {
          return (
            <Card className="centered">
              <Details {...displayMethod.data} />

              <Divider />

              <Flex justify="center">
                <DisplayMethodCard metadata={displayMethodMetadata.data} />
              </Flex>

              <Divider />

              <Space direction="vertical">
                <Typography.Text type="secondary">DISPLAY METHOD METADATA</Typography.Text>

                <Detail label="Title" text={displayMethodMetadata.data.title} />
                <Detail label="Description" text={displayMethodMetadata.data.description} />
                <Detail label="Issuer name" text={displayMethodMetadata.data.issuerName} />
                <Detail label="Title color" text={displayMethodMetadata.data.titleTextColor} />
                <Detail
                  label="Description color"
                  text={displayMethodMetadata.data.descriptionTextColor}
                />
                <Detail
                  label="Issuer name color"
                  text={displayMethodMetadata.data.issuerTextColor}
                />

                <Detail
                  copyable
                  href={displayMethodMetadata.data.backgroundImageUrl}
                  label="Background image URL"
                  text={displayMethodMetadata.data.backgroundImageUrl}
                />

                <Detail
                  copyable
                  href={displayMethodMetadata.data.logo.uri}
                  label="Logo URL"
                  text={displayMethodMetadata.data.logo.uri}
                />

                <Detail label="Logo alt" text={displayMethodMetadata.data.logo.alt} />
              </Space>
            </Card>
          );
        }
      })()}
    </SiderLayoutContent>
  );
}
