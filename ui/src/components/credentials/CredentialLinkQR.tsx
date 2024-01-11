import { Avatar, Button, Space, Typography, message } from "antd";
import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";

import {
  AuthQRCode,
  ImportQRCode,
  createAuthQRCode,
  getImportQRCode,
} from "src/adapters/api/credentials";
import AlertIcon from "src/assets/icons/alert-circle.svg?react";
import CheckIcon from "src/assets/icons/check.svg?react";
import QRIcon from "src/assets/icons/qr-code.svg?react";
import IconRefresh from "src/assets/icons/refresh-ccw-01.svg?react";
import { ClaimCredentialModal } from "src/components/credentials/ClaimCredentialModal";
import { CredentialQR } from "src/components/credentials/CredentialQR";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { useEnvContext } from "src/contexts/Env";
import { AppError } from "src/domain";
import { AsyncTask, hasAsyncTaskFailed, isAsyncTaskDataAvailable } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { POLLING_INTERVAL } from "src/utils/constants";

const PUSH_NOTIFICATIONS_REMINDER =
  "Please ensure that you have enabled push notifications on your wallet app.";

export function CredentialLinkQR() {
  const env = useEnvContext();

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [authQRCode, setAuthQRCode] = useState<AsyncTask<AuthQRCode, AppError>>({
    status: "pending",
  });
  const [importQRCheck, setImportQRCheck] = useState<AsyncTask<ImportQRCode, AppError>>({
    status: "pending",
  });

  const [messageAPI, messageContext] = message.useMessage();
  const { linkID } = useParams();

  const createCredentialQR = useCallback(
    async (signal: AbortSignal) => {
      if (linkID) {
        setAuthQRCode({ status: "loading" });

        const response = await createAuthQRCode({ env, linkID, signal });

        if (response.success) {
          setAuthQRCode({ data: response.data, status: "successful" });
        } else {
          if (!isAbortedError(response.error)) {
            setAuthQRCode({ error: response.error, status: "failed" });
          }
        }
      }
    },
    [linkID, env]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(createCredentialQR);

    return aborter;
  }, [createCredentialQR]);

  useEffect(() => {
    const checkCredentialQRCode = async () => {
      if (isAsyncTaskDataAvailable(authQRCode) && linkID) {
        const response = await getImportQRCode({
          env,
          linkID,
          sessionID: authQRCode.data.sessionID,
        });

        if (response.success) {
          if (response.data.status !== "pending") {
            setImportQRCheck({ data: response.data, status: "successful" });

            const { proofTypes } = authQRCode.data.linkDetail;

            if (proofTypes.includes("MTP")) {
              void messageAPI.info("Issuance process started");
            }

            if (proofTypes.includes("SIG")) {
              void messageAPI.success("Credential sent");
            }
          }
        } else {
          setImportQRCheck({ error: response.error, status: "failed" });

          void messageAPI.error(response.error.message);
        }
      }
    };

    const checkQRCredentialStatusTimer = setInterval(() => {
      if (
        (isAsyncTaskDataAvailable(importQRCheck) && importQRCheck.data.status !== "pending") ||
        hasAsyncTaskFailed(importQRCheck)
      ) {
        clearInterval(checkQRCredentialStatusTimer);
      } else {
        void checkCredentialQRCode();
      }
    }, POLLING_INTERVAL);

    return () => clearInterval(checkQRCredentialStatusTimer);
  }, [authQRCode, env, importQRCheck, linkID, messageAPI]);

  const onStartAgain = () => {
    makeRequestAbortable(createCredentialQR);
    setImportQRCheck({ status: "pending" });
  };

  const appError = hasAsyncTaskFailed(authQRCode)
    ? authQRCode.error
    : hasAsyncTaskFailed(importQRCheck)
      ? importQRCheck.error
      : undefined;

  return (
    <>
      {messageContext}

      {(() => {
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
          if (isAsyncTaskDataAvailable(importQRCheck) && importQRCheck.data.status !== "pending") {
            const { proofTypes } = authQRCode.data.linkDetail;

            if (proofTypes.length > 1) {
              return (
                <>
                  <Space align="center" direction="vertical" size="large">
                    <Avatar className="avatar-color-success" icon={<CheckIcon />} size={56} />

                    <Typography.Title level={2}>
                      Credential sent via notification. On-chain capabilities are pending.
                    </Typography.Title>

                    <Typography.Text style={{ fontSize: 18 }} type="secondary">
                      You will receive an additional version of the credential containing an MTP
                      proof.
                      <br />
                      {PUSH_NOTIFICATIONS_REMINDER}
                    </Typography.Text>

                    <Button onClick={() => setIsModalOpen(true)} type="link">
                      Missed the notification?
                    </Button>

                    {isModalOpen && importQRCheck.data.qrCode && (
                      <ClaimCredentialModal
                        onClose={() => setIsModalOpen(false)}
                        qrCode={importQRCheck.data.qrCode}
                      />
                    )}
                  </Space>
                </>
              );
            }

            return proofTypes[0] === "SIG" ? (
              <>
                <Space align="center" direction="vertical" size="large">
                  <Avatar className="avatar-color-success" icon={<CheckIcon />} size={56} />

                  <Typography.Title level={2}>Credential sent via notification</Typography.Title>

                  <Button onClick={() => setIsModalOpen(true)} type="link">
                    Missed the notification?
                  </Button>

                  {isModalOpen && importQRCheck.data.qrCode && (
                    <ClaimCredentialModal
                      onClose={() => setIsModalOpen(false)}
                      qrCode={importQRCheck.data.qrCode}
                    />
                  )}
                </Space>
              </>
            ) : (
              <>
                <Space align="center" direction="vertical" size="large">
                  <Avatar className="avatar-color-success" icon={<CheckIcon />} size={56} />

                  <Typography.Title level={2}>
                    You will receive your credential via a notification
                  </Typography.Title>

                  <Typography.Text style={{ fontSize: 18 }} type="secondary">
                    {PUSH_NOTIFICATIONS_REMINDER}
                  </Typography.Text>

                  <Button icon={<IconRefresh />} onClick={onStartAgain}>
                    Start again
                  </Button>
                </Space>
              </>
            );
          }

          return (
            <CredentialQR
              qrCodeLink={authQRCode.data.qrCode}
              qrCodeRaw={authQRCode.data.qrCode}
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
      })()}
    </>
  );
}
