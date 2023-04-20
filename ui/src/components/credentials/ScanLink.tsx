import { Avatar, Button, Card, Col, Grid, Image, Row, Space, Typography, message } from "antd";
import { QRCodeSVG } from "qrcode.react";
import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";

import { APIError, HTTPStatusError } from "src/adapters/api";
import {
  CredentialQRCheck,
  ShareCredentialQRCode,
  createCredentialLinkQRCode,
  getCredentialLinkQRCode,
} from "src/adapters/api/credentials";
import { ReactComponent as AlertIcon } from "src/assets/icons/alert-circle.svg";
import { ReactComponent as QRIcon } from "src/assets/icons/qr-code.svg";
import { ReactComponent as IconRefresh } from "src/assets/icons/refresh-ccw-01.svg";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { useEnvContext } from "src/contexts/env";
import { AsyncTask, hasAsyncTaskFailed, isAsyncTaskDataAvailable } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import {
  IMAGE_PLACEHOLDER_PATH,
  QR_CODE_POLLING_INTERVAL,
  WALLET_APP_STORE_URL,
  WALLET_PLAY_STORE_URL,
} from "src/utils/constants";

export function ScanLink() {
  const env = useEnvContext();
  const [shareCredentialQRCode, setShareCredentialQRCode] = useState<
    AsyncTask<ShareCredentialQRCode, APIError>
  >({
    status: "pending",
  });
  const [credentialQRCheck, setCredentialQRCheck] = useState<
    AsyncTask<CredentialQRCheck, APIError>
  >({
    status: "pending",
  });

  const { lg } = Grid.useBreakpoint();
  const { linkID } = useParams();

  const createCredentialQR = useCallback(
    async (signal: AbortSignal) => {
      if (linkID) {
        setShareCredentialQRCode({ status: "loading" });

        const response = await createCredentialLinkQRCode({ env, linkID, signal });

        if (response.isSuccessful) {
          setShareCredentialQRCode({ data: response.data, status: "successful" });
        } else {
          if (!isAbortedError(response.error)) {
            setShareCredentialQRCode({ error: response.error, status: "failed" });
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
      if (isAsyncTaskDataAvailable(shareCredentialQRCode) && linkID) {
        const response = await getCredentialLinkQRCode({
          env,
          linkID,
          sessionID: shareCredentialQRCode.data.sessionID,
        });

        if (response.isSuccessful) {
          setCredentialQRCheck({ data: response.data, status: "successful" });

          if (response.data.status === "done") {
            void message.success("Credential successfully shared");
          }
        } else {
          setCredentialQRCheck({ error: response.error, status: "failed" });

          void message.error(response.error.message);
        }
      }
    };

    const checkQRCredentialStatusTimer = setInterval(() => {
      if (
        isAsyncTaskDataAvailable(credentialQRCheck) &&
        credentialQRCheck.data.status === "pending"
      ) {
        void checkCredentialQRCode();
      } else {
        clearInterval(checkQRCredentialStatusTimer);
      }
    }, QR_CODE_POLLING_INTERVAL);

    return () => clearInterval(checkQRCredentialStatusTimer);
  }, [shareCredentialQRCode, linkID, credentialQRCheck, env]);

  const onStartAgain = () => {
    makeRequestAbortable(createCredentialQR);
    setCredentialQRCheck({ status: "pending" });
  };

  const hasFailed = hasAsyncTaskFailed(shareCredentialQRCode)
    ? shareCredentialQRCode.error
    : hasAsyncTaskFailed(credentialQRCheck)
    ? credentialQRCheck.error
    : undefined;

  if (hasFailed) {
    if (hasFailed.status === HTTPStatusError.NotFound) {
      return (
        <Space align="center" direction="vertical" size="large">
          <Avatar className="avatar-color-error" icon={<QRIcon />} size={56} />

          <Typography.Title level={2}>Claim link is invalid</Typography.Title>

          <Typography.Text type="secondary">
            In case you think this is an error, please contact the issuer of this claim.
          </Typography.Text>
        </Space>
      );
    } else if (hasFailed.status === HTTPStatusError.BadRequest) {
      return (
        <Space align="center" direction="vertical" size="large">
          <Avatar className="avatar-color-error" icon={<AlertIcon />} size={56} />

          <Typography.Title level={2}>QR code expired, start again</Typography.Title>

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

  return !isAsyncTaskDataAvailable(shareCredentialQRCode) ? (
    <LoadingResult />
  ) : (
    <Space align="center" direction="vertical" size="large">
      <Avatar
        shape="square"
        size={64}
        src={shareCredentialQRCode.data.issuer.logo || IMAGE_PLACEHOLDER_PATH}
      />

      <Space
        direction="vertical"
        style={{ padding: "0 24px", textAlign: "center", width: lg ? 800 : "100%" }}
      >
        <Typography.Title level={2}>
          {isAsyncTaskDataAvailable(credentialQRCheck) && credentialQRCheck.data.status === "done"
            ? "Please scan to add credential to your wallet"
            : `You received a credential from ${shareCredentialQRCode.data.issuer.displayName}`}
        </Typography.Title>

        <Typography.Text style={{ fontSize: 18 }} type="secondary">
          {isAsyncTaskDataAvailable(credentialQRCheck) && credentialQRCheck.data.status === "done"
            ? "If you already received a push notification and added the claim to your mobile device, please disregard this message."
            : "Scan the QR code with your Polygon ID wallet to accept it."}
        </Typography.Text>
      </Space>

      {isAsyncTaskDataAvailable(credentialQRCheck) && credentialQRCheck.data.status === "done" ? (
        <Button icon={<IconRefresh />} onClick={onStartAgain} type="link">
          Start again
        </Button>
      ) : (
        <Space>
          <Typography.Link href={WALLET_APP_STORE_URL} target="_blank">
            <Image preview={false} src="/images/apple-store.svg" />
          </Typography.Link>

          <Typography.Link href={WALLET_PLAY_STORE_URL} target="_blank">
            <Image preview={false} src="/images/google-play.svg" />
          </Typography.Link>
        </Space>
      )}

      <Card bodyStyle={{ padding: 0 }} style={{ margin: "auto", width: lg ? 800 : "100%" }}>
        <Row>
          <Col
            className="full-width"
            style={{
              background: `url("/images/noise-bg.png"), linear-gradient(50deg, ${
                isAsyncTaskDataAvailable(credentialQRCheck) &&
                credentialQRCheck.data.status === "done"
                  ? "rgb(255 152 57) 0%, rgba(255, 214, 174, 1) 50%"
                  : "rgb(130 101 208) 0%, rgba(221, 178, 248, 1) 50%"
              })`,
              borderRadius: 8,
              padding: 24,
            }}
          >
            <QRCodeSVG
              className="full-width"
              includeMargin
              level="H"
              style={{ height: 300 }}
              value={
                isAsyncTaskDataAvailable(credentialQRCheck) &&
                credentialQRCheck.data.status === "done"
                  ? JSON.stringify(credentialQRCheck.data.qrCode)
                  : JSON.stringify(shareCredentialQRCode.data.qrCode)
              }
            />
          </Col>
        </Row>
        <Row>
          <Col
            style={{
              padding: 24,
            }}
          >
            <Space direction="vertical" size="large">
              <Typography.Title ellipsis={{ tooltip: true }} level={3}>
                {shareCredentialQRCode.data.linkDetail.schemaType}
              </Typography.Title>
              <Typography.Title level={5} type="secondary">
                Attributes
              </Typography.Title>

              {/* TODO Credentials epic: PID-601 */}
              {/* {shareCredentialQRCode.data.linkDetail.attributes.map((attribute) => {
                const formattedValue = formatAttributeValue(
                  attribute,
                  shareCredentialQRCode.data.offerDetails.schemaTemplate.attributes
                );

                return (
                  <Space direction="vertical" key={attribute.name}>
                    <Typography.Text ellipsis={{ tooltip: true }} type="secondary">
                      {attribute.name}
                    </Typography.Text>

                    <Typography.Text strong>
                      {formattedValue.success ? formattedValue.data : formattedValue.error}
                    </Typography.Text>
                  </Space>
                );
              })} */}
            </Space>
          </Col>
        </Row>
      </Card>
    </Space>
  );
}
