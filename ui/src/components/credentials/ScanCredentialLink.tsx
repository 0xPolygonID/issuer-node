import { Avatar, Button, Card, Col, Grid, Image, Row, Space, Typography, message } from "antd";
import { QRCodeSVG } from "qrcode.react";
import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";

import {
  CredentialQRCheck,
  CredentialQRStatus,
  ShareCredentialQRCode,
  credentialsQRCheck,
  credentialsQRCreate,
} from "src/adapters/api/credentials";
import { formatAttributeValue } from "src/adapters/parsers/forms";
import { ReactComponent as QRIcon } from "src/assets/icons/qr-code.svg";
import { ReactComponent as IconRefresh } from "src/assets/icons/refresh-ccw-01.svg";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { useEnvContext } from "src/contexts/env";
import { APIError, HTTPStatusError } from "src/utils/adapters";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import {
  IMAGE_PLACEHOLDER_PATH,
  QR_CODE_POLLING_INTERVAL,
  WALLET_APP_STORE_URL,
  WALLET_PLAY_STORE_URL,
} from "src/utils/constants";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/types";

export function ScanCredentialLink() {
  const env = useEnvContext();
  const [shareCredentialQRCode, setShareCredentialQRCode] = useState<
    AsyncTask<ShareCredentialQRCode, APIError>
  >({
    status: "pending",
  });
  const [credentialQRCheck, setCredentialQRCheck] = useState<CredentialQRCheck>({
    status: CredentialQRStatus.Pending,
  });

  const { lg } = Grid.useBreakpoint();
  const { credentialID } = useParams();

  const createCredentialQR = useCallback(
    async (signal: AbortSignal) => {
      if (credentialID) {
        setShareCredentialQRCode({ status: "loading" });

        const response = await credentialsQRCreate({ id: credentialID, signal });

        if (response.isSuccessful) {
          setShareCredentialQRCode({ data: response.data, status: "successful" });
        } else {
          if (!isAbortedError(response.error)) {
            setShareCredentialQRCode({ error: response.error, status: "failed" });
          }
        }
      }
    },
    [credentialID]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(createCredentialQR);

    return aborter;
  }, [createCredentialQR]);

  useEffect(() => {
    const checkCredentialQRCode = async () => {
      if (isAsyncTaskDataAvailable(shareCredentialQRCode) && credentialID) {
        const response = await credentialsQRCheck({
          credentialID,
          sessionID: shareCredentialQRCode.data.sessionID,
        });

        if (response.isSuccessful) {
          setCredentialQRCheck(response.data);
          if (response.data.status === CredentialQRStatus.Done) {
            void message.success("Credential successfully shared");
          }
        } else {
          setCredentialQRCheck({ status: CredentialQRStatus.Error });
        }
      }
    };

    const checkQRCredentialStatusTimer = setInterval(() => {
      if (credentialQRCheck.status === CredentialQRStatus.Pending) {
        void checkCredentialQRCode();
      } else {
        clearInterval(checkQRCredentialStatusTimer);
      }
    }, QR_CODE_POLLING_INTERVAL);

    return () => clearInterval(checkQRCredentialStatusTimer);
  }, [shareCredentialQRCode, credentialID, credentialQRCheck, env]);

  const onStartAgain = () => {
    makeRequestAbortable(createCredentialQR);
    setCredentialQRCheck({ status: CredentialQRStatus.Pending });
  };

  if (shareCredentialQRCode.status === "failed") {
    if (shareCredentialQRCode.error.status === HTTPStatusError.NotFound) {
      return (
        <Space
          align="center"
          direction="vertical"
          size="large"
          style={{ textAlign: "center", width: 400 }}
        >
          <Avatar className="avatar-color-primary" icon={<QRIcon />} size={56} />

          <Typography.Title level={2}>Credential link is invalid</Typography.Title>

          <Typography.Text type="secondary">
            In case you think this is an error, please contact the issuer of this credential.
          </Typography.Text>
        </Space>
      );
    }
    return (
      <ErrorResult
        error={shareCredentialQRCode.error.message}
        labelRetry="Start again"
        onRetry={onStartAgain}
      />
    );
  }

  return isAsyncTaskStarting(shareCredentialQRCode) ? (
    <LoadingResult />
  ) : (
    <Space align="center" direction="vertical" size="large">
      <Avatar
        shape="square"
        size={64}
        src={shareCredentialQRCode.data.issuer.logo || IMAGE_PLACEHOLDER_PATH}
      />

      <Space direction="vertical" style={{ textAlign: "center" }}>
        <Typography.Title level={2}>
          {credentialQRCheck.status === CredentialQRStatus.Done
            ? "Scan again to add the credential to your wallet"
            : `You received a credential request from ${shareCredentialQRCode.data.issuer.displayName}`}
        </Typography.Title>

        <Typography.Text>
          {credentialQRCheck.status === CredentialQRStatus.Done
            ? "If you already received a push notification and added the credential to your mobile device, please disregard this message."
            : "Scan the QR code with your Polygon ID wallet to add the credential to your wallet."}
        </Typography.Text>
      </Space>

      {credentialQRCheck.status === CredentialQRStatus.Done && (
        <Button icon={<IconRefresh />} onClick={onStartAgain} type="link">
          Start again
        </Button>
      )}

      <Card bodyStyle={{ padding: 0 }} style={{ margin: "auto", width: lg ? 800 : "100%" }}>
        <Row>
          <Col
            className="full-width"
            md={13}
            style={{
              background:
                'url("/images/noise-bg.png"), linear-gradient(50deg, rgb(130 101 208) 15%, rgba(221, 178, 248, 1) 44%)',
              borderRadius: 8,
              padding: 24,
            }}
          >
            <QRCodeSVG
              className="full-width"
              includeMargin
              level="H"
              style={{ height: "100%" }}
              value={
                credentialQRCheck.status === CredentialQRStatus.Done
                  ? JSON.stringify(credentialQRCheck.qrcode)
                  : JSON.stringify(shareCredentialQRCode.data.qrcode)
              }
            />
          </Col>

          <Col
            md={11}
            style={{
              padding: 24,
            }}
          >
            <Space direction="vertical" size="large" style={{ maxWidth: "50vw" }}>
              <Typography.Title ellipsis={{ tooltip: true }} level={3}>
                {shareCredentialQRCode.data.offerDetails.schemaTemplate.type}
              </Typography.Title>

              <Typography.Title level={5} type="secondary">
                Attributes
              </Typography.Title>

              {shareCredentialQRCode.data.offerDetails.attributeValues.map((attribute) => {
                const formattedValue = formatAttributeValue(
                  attribute,
                  //TODO Credentials epic
                  // shareCredentialQRCode.data.offerDetails.schemaTemplate.attributes
                  []
                );

                return (
                  <Space direction="vertical" key={attribute.attributeKey}>
                    <Typography.Text ellipsis={{ tooltip: true }} type="secondary">
                      {attribute.attributeKey}
                    </Typography.Text>

                    <Typography.Text strong>
                      {formattedValue.success ? formattedValue.data : formattedValue.error}
                    </Typography.Text>
                  </Space>
                );
              })}
            </Space>
          </Col>
        </Row>
      </Card>

      <Space align="center" direction="vertical">
        <Typography.Text type="secondary">Download Polygon ID wallet app</Typography.Text>

        <Space>
          <Typography.Link href={WALLET_APP_STORE_URL} target="_blank">
            <Image preview={false} src="/images/apple-store.svg" />
          </Typography.Link>

          <Typography.Link href={WALLET_PLAY_STORE_URL} target="_blank">
            <Image preview={false} src="/images/google-play.svg" />
          </Typography.Link>
        </Space>
      </Space>
    </Space>
  );
}
