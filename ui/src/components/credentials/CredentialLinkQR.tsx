import { Avatar, Button, Space, Typography, message } from "antd";
import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";

import { APIError } from "src/adapters/api";
import {
  AuthQRCode,
  ImportQRCode,
  createAuthQRCode,
  getImportQRCode,
} from "src/adapters/api/credentials";
import { ReactComponent as AlertIcon } from "src/assets/icons/alert-circle.svg";
import { ReactComponent as CheckIcon } from "src/assets/icons/check.svg";
import { ReactComponent as QRIcon } from "src/assets/icons/qr-code.svg";
import { ReactComponent as IconRefresh } from "src/assets/icons/refresh-ccw-01.svg";
import { ClaimCredentialModal } from "src/components/credentials/ClaimCredentialModal";
import { CredentialQR } from "src/components/credentials/CredentialQR";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { useEnvContext } from "src/contexts/Env";
import { AsyncTask, hasAsyncTaskFailed, isAsyncTaskDataAvailable } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { POLLING_INTERVAL } from "src/utils/constants";

export function CredentialLinkQR() {
  const env = useEnvContext();

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [authQRCode, setAuthQRCode] = useState<AsyncTask<AuthQRCode, APIError>>({
    status: "pending",
  });
  const [importQRCheck, setImportQRCheck] = useState<AsyncTask<ImportQRCode, APIError>>({
    status: "pending",
  });

  const { linkID } = useParams();

  const createCredentialQR = useCallback(
    async (signal: AbortSignal) => {
      if (linkID) {
        setAuthQRCode({ status: "loading" });

        const response = await createAuthQRCode({ env, linkID, signal });

        if (response.isSuccessful) {
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

        if (response.isSuccessful) {
          if (response.data.status !== "pending") {
            setImportQRCheck({ data: response.data, status: "successful" });

            const { proofTypes } = authQRCode.data.linkDetail;

            if (proofTypes.includes("MTP")) {
              void message.info("Issuance process started");
            }

            if (proofTypes.includes("SIG")) {
              void message.success("Credential sent");
            }
          }
        } else {
          setImportQRCheck({ error: response.error, status: "failed" });

          void message.error(response.error.message);
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
  }, [authQRCode, linkID, importQRCheck, env]);

  const onStartAgain = () => {
    makeRequestAbortable(createCredentialQR);
    setImportQRCheck({ status: "pending" });
  };

  const hasFailed = hasAsyncTaskFailed(authQRCode)
    ? authQRCode.error
    : hasAsyncTaskFailed(importQRCheck)
    ? importQRCheck.error
    : undefined;

  if (hasFailed) {
    if (hasFailed.status === 404) {
      return (
        <Space align="center" direction="vertical" size="large">
          <Avatar className="avatar-color-error" icon={<QRIcon />} size={56} />

          <Typography.Title level={2}>Credential link is invalid</Typography.Title>

          <Typography.Text type="secondary">
            In case you think this is an error, please contact the issuer of this claim.
          </Typography.Text>
        </Space>
      );
    } else if (hasFailed.status === 400) {
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
      <ErrorResult error={hasFailed.message} labelRetry="Start again" onRetry={onStartAgain} />
    );
  }

  if (!isAsyncTaskDataAvailable(authQRCode)) {
    return <LoadingResult />;
  }

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
              You will receive an additional version of the credential containing an MTP proof.
              <Typography.Paragraph style={{ fontSize: 18, textAlign: "center" }} type="secondary">
                Please ensure that you have enabled push notifications on the application.
              </Typography.Paragraph>
            </Typography.Text>

            <Button onClick={() => setIsModalOpen(true)} type="link">
              Missed the notification?
            </Button>

            {isModalOpen && (
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

          {isModalOpen && (
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
            Please ensure that you have enabled push notifications on the application.
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
      qrCode={authQRCode.data.qrCode}
      schemaType={authQRCode.data.linkDetail.schemaType}
      subTitle="Scan the QR code with your Polygon ID wallet to accept it. Make sure push notifications are enabled."
    />
  );
}
