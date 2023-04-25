import { Avatar, Button, Card, Col, Grid, Image, Row, Space, Typography, message } from "antd";
import { QRCodeSVG } from "qrcode.react";
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
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { useEnvContext } from "src/contexts/Env";
import { AsyncTask, hasAsyncTaskFailed, isAsyncTaskDataAvailable } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import {
  IMAGE_PLACEHOLDER_PATH,
  POLLING_INTERVAL,
  WALLET_APP_STORE_URL,
  WALLET_PLAY_STORE_URL,
} from "src/utils/constants";

export function CredentialLinkQR() {
  const env = useEnvContext();

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [authQRCode, setAuthQRCode] = useState<AsyncTask<AuthQRCode, APIError>>({
    status: "pending",
  });
  const [importQRCheck, setImportQRCheck] = useState<AsyncTask<ImportQRCode, APIError>>({
    status: "pending",
  });

  const { lg } = Grid.useBreakpoint();
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

            const { proofType } = authQRCode.data.linkDetail;

            if (proofType === "MTP" || proofType === "MTP & SIG") {
              void message.info("Issuance process started");
            }

            if (proofType === "SIG" || proofType === "MTP & SIG") {
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
    const { proofType } = authQRCode.data.linkDetail;

    switch (proofType) {
      case "MTP & SIG": {
        return (
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
        );
      }
      case "MTP": {
        return (
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
        );
      }
      case "SIG": {
        return (
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
        );
      }
    }
  }

  return (
    <Space align="center" direction="vertical" size="large">
      <Avatar
        shape="square"
        size={64}
        src={authQRCode.data.issuer.logo || IMAGE_PLACEHOLDER_PATH}
      />

      <Space
        direction="vertical"
        style={{ padding: "0 24px", textAlign: "center", width: lg ? 800 : "100%" }}
      >
        <Typography.Title level={2}>
          {authQRCode.data.issuer.displayName} wants to send you a credential
        </Typography.Title>

        <Typography.Text style={{ fontSize: 18 }} type="secondary">
          Scan the QR code with your Polygon ID wallet to accept it. Make sure push notifications
          are enabled
        </Typography.Text>
      </Space>

      <Space>
        <Typography.Link href={WALLET_APP_STORE_URL} target="_blank">
          <Image preview={false} src="/images/apple-store.svg" />
        </Typography.Link>

        <Typography.Link href={WALLET_PLAY_STORE_URL} target="_blank">
          <Image preview={false} src="/images/google-play.svg" />
        </Typography.Link>
      </Space>

      <Card bodyStyle={{ padding: 0 }} style={{ margin: "auto", width: lg ? 800 : "100%" }}>
        <Row>
          <Col
            className="full-width"
            style={{
              background:
                'url("/images/noise-bg.png"), linear-gradient(50deg, rgb(130 101 208) 0%, rgba(221, 178, 248, 1) 50%',
              borderRadius: 8,
              padding: 24,
            }}
          >
            <QRCodeSVG
              className="full-width"
              includeMargin
              level="H"
              style={{ height: 300 }}
              value={JSON.stringify(authQRCode.data.qrCode)}
            />
          </Col>
        </Row>
        <Row>
          <Col
            style={{
              padding: 24,
              paddingBottom: 8,
            }}
          >
            <Typography.Title ellipsis={{ tooltip: true }} level={3}>
              {authQRCode.data.linkDetail.schemaType}
            </Typography.Title>
          </Col>
        </Row>
      </Card>
    </Space>
  );
}
