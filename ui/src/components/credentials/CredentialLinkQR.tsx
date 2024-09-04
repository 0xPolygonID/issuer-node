import { Avatar, Button, Space, Typography } from "antd";
import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";

import { AuthQRCode, createAuthQRCode } from "src/adapters/api/credentials";
import AlertIcon from "src/assets/icons/alert-circle.svg?react";
import QRIcon from "src/assets/icons/qr-code.svg?react";
import IconRefresh from "src/assets/icons/refresh-ccw-01.svg?react";
import { CredentialQR } from "src/components/credentials/CredentialQR";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { useEnvContext } from "src/contexts/Env";
import { useIssuerContext } from "src/contexts/Issuer";
import { AppError } from "src/domain";
import { AsyncTask, hasAsyncTaskFailed, isAsyncTaskDataAvailable } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";

const PUSH_NOTIFICATIONS_REMINDER =
  "Please ensure that you have enabled push notifications on your wallet app.";

export function CredentialLinkQR() {
  const env = useEnvContext();
  const { issuerIdentifier } = useIssuerContext();

  const [authQRCode, setAuthQRCode] = useState<AsyncTask<AuthQRCode, AppError>>({
    status: "pending",
  });

  const { linkID } = useParams();

  const createCredentialQR = useCallback(
    async (signal: AbortSignal) => {
      if (linkID) {
        setAuthQRCode({ status: "loading" });

        const response = await createAuthQRCode({ env, issuerIdentifier, linkID, signal });

        if (response.success) {
          setAuthQRCode({ data: response.data, status: "successful" });
        } else {
          if (!isAbortedError(response.error)) {
            setAuthQRCode({ error: response.error, status: "failed" });
          }
        }
      }
    },
    [linkID, env, issuerIdentifier]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(createCredentialQR);

    return aborter;
  }, [createCredentialQR]);

  const onStartAgain = () => {
    makeRequestAbortable(createCredentialQR);
  };

  const appError = hasAsyncTaskFailed(authQRCode) ? authQRCode.error : undefined;

  return (() => {
    if (appError) {
      if (appError.type === "request-error" && appError.error.response?.status === 404) {
        return (
          <Space align="center" direction="vertical" size="large">
            <Avatar className="avatar-color-error" icon={<QRIcon />} size={56} />

            <Typography.Title level={2}>Credential link is invalid</Typography.Title>

            <Typography.Text type="secondary">
              If you think this is an error, please contact the issuer of this credential.
            </Typography.Text>
          </Space>
        );
      } else if (appError.type === "request-error" && appError.error.response?.status === 400) {
        return (
          <Space align="center" direction="vertical" size="large">
            <Avatar className="avatar-color-error" icon={<AlertIcon />} size={56} />

            <Typography.Title level={2}>
              The credential link has expired, please start again
            </Typography.Title>

            <Button icon={<IconRefresh />} onClick={onStartAgain} type="link">
              Start again
            </Button>
          </Space>
        );
      }

      return (
        <ErrorResult error={appError.message} labelRetry="Start again" onRetry={onStartAgain} />
      );
    }

    if (!isAsyncTaskDataAvailable(authQRCode)) {
      return <LoadingResult />;
    } else {
      return (
        <CredentialQR
          qrCodeLink={authQRCode.data.deepLink}
          qrCodeRaw={authQRCode.data.qrCodeRaw}
          schemaType={authQRCode.data.linkDetail.schemaType}
          subTitle={
            <>
              Scan the QR code with your Polygon ID wallet to accept it.
              <br />
              {PUSH_NOTIFICATIONS_REMINDER}
            </>
          }
        />
      );
    }
  })();
}
