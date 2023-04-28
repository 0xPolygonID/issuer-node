import { Card, Col, Grid, Image, Row, Space, Typography } from "antd";
import { QRCodeSVG } from "qrcode.react";
import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";

import { APIError } from "src/adapters/api";
import { getIssuedQRCode } from "src/adapters/api/credentials";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { useEnvContext } from "src/contexts/Env";
import { IssuedQRCode } from "src/domain";
import { AsyncTask, hasAsyncTaskFailed, isAsyncTaskDataAvailable } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { WALLET_APP_STORE_URL, WALLET_PLAY_STORE_URL } from "src/utils/constants";

export function CredentialIssuedQR() {
  const env = useEnvContext();
  const [issuedQRCode, setIssuedQRCode] = useState<AsyncTask<IssuedQRCode, APIError>>({
    status: "pending",
  });

  const { lg } = Grid.useBreakpoint();
  const { credentialID } = useParams();

  const createCredentialQR = useCallback(
    async (signal: AbortSignal) => {
      if (credentialID) {
        setIssuedQRCode({ status: "loading" });

        const response = await getIssuedQRCode({ credentialID, env, signal });

        if (response.isSuccessful) {
          setIssuedQRCode({ data: response.data, status: "successful" });
        } else {
          if (!isAbortedError(response.error)) {
            setIssuedQRCode({ error: response.error, status: "failed" });
          }
        }
      }
    },
    [credentialID, env]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(createCredentialQR);

    return aborter;
  }, [createCredentialQR]);

  const onStartAgain = () => {
    makeRequestAbortable(createCredentialQR);
    setIssuedQRCode({ status: "pending" });
  };

  const hasFailed = hasAsyncTaskFailed(issuedQRCode) ? issuedQRCode.error : undefined;

  if (hasFailed) {
    return (
      <ErrorResult error={hasFailed.message} labelRetry="Start again" onRetry={onStartAgain} />
    );
  }

  if (!isAsyncTaskDataAvailable(issuedQRCode)) {
    return <LoadingResult />;
  }

  return (
    <Space align="center" direction="vertical" size="large">
      <Space
        direction="vertical"
        style={{ padding: "0 24px", textAlign: "center", width: lg ? 800 : "100%" }}
      >
        <Typography.Title level={3}>
          Scan the QR code to add the credential to your wallet
        </Typography.Title>
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
              value={JSON.stringify(issuedQRCode.data.qrCode)}
            />
          </Col>
        </Row>
        {issuedQRCode.data.schemaType && (
          <Row>
            <Col
              style={{
                padding: 24,
                paddingBottom: 8,
              }}
            >
              <Typography.Title ellipsis={{ tooltip: true }} level={3}>
                {issuedQRCode.data.schemaType}
              </Typography.Title>
            </Col>
          </Row>
        )}
      </Card>
    </Space>
  );
}
