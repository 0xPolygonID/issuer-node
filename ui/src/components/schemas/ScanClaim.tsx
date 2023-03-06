import { Avatar, Button, Card, Col, Grid, Image, Row, Space, Typography, message } from "antd";
import { QRCodeSVG } from "qrcode.react";
import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";

import {
  ClaimQRCheck,
  ClaimQRStatus,
  ShareClaimQRCode,
  claimsQRCheck,
  claimsQRCreate,
} from "src/adapters/api/claims";
import { formatAttributeValue } from "src/adapters/parsers/forms";
import { ReactComponent as QRIcon } from "src/assets/icons/qr-code.svg";
import { ReactComponent as IconRefresh } from "src/assets/icons/refresh-ccw-01.svg";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { APIError, HTTPStatusError } from "src/utils/adapters";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import {
  IMAGE_PLACEHOLDER_PATH,
  QR_CODE_POLLING_INTERVAL,
  WALLET_APP_STORE_URL,
  WALLET_PLAY_STORE_URL,
} from "src/utils/constants";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/types";

export function ScanClaim() {
  const [shareClaimQRCode, setShareClaimQRCode] = useState<AsyncTask<ShareClaimQRCode, APIError>>({
    status: "pending",
  });
  const [claimQRCheck, setClaimQRCheck] = useState<ClaimQRCheck>({ status: ClaimQRStatus.Pending });

  const { lg } = Grid.useBreakpoint();
  const { claimID } = useParams();

  const createClaimQR = useCallback(
    async (signal: AbortSignal) => {
      if (claimID) {
        setShareClaimQRCode({ status: "loading" });
        const response = await claimsQRCreate({ id: claimID, signal });
        if (response.isSuccessful) {
          setShareClaimQRCode({ data: response.data, status: "successful" });
        } else {
          if (!isAbortedError(response.error)) {
            setShareClaimQRCode({ error: response.error, status: "failed" });
          }
        }
      }
    },
    [claimID]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(createClaimQR);
    return aborter;
  }, [createClaimQR]);

  useEffect(() => {
    const checkClaimQRCode = async () => {
      if (isAsyncTaskDataAvailable(shareClaimQRCode) && claimID) {
        const response = await claimsQRCheck({
          claimID,
          sessionID: shareClaimQRCode.data.sessionID,
        });

        if (response.isSuccessful) {
          setClaimQRCheck(response.data);
          if (response.data.status === ClaimQRStatus.Done) {
            void message.success("Claim successfully shared");
          }
        } else {
          setClaimQRCheck({ status: ClaimQRStatus.Error });
        }
      }
    };

    const checkQRClaimStatusTimer = setInterval(() => {
      if (claimQRCheck.status === ClaimQRStatus.Pending) {
        void checkClaimQRCode();
      } else {
        clearInterval(checkQRClaimStatusTimer);
      }
    }, QR_CODE_POLLING_INTERVAL);
    return () => clearInterval(checkQRClaimStatusTimer);
  }, [shareClaimQRCode, claimID, claimQRCheck]);

  const onStartAgain = () => {
    makeRequestAbortable(createClaimQR);
    setClaimQRCheck({ status: ClaimQRStatus.Pending });
  };

  if (shareClaimQRCode.status === "failed") {
    if (shareClaimQRCode.error.status === HTTPStatusError.NotFound) {
      return (
        <Space
          align="center"
          direction="vertical"
          size="large"
          style={{ textAlign: "center", width: 400 }}
        >
          <Avatar className="avatar-color-primary" icon={<QRIcon />} size={56} />

          <Typography.Title level={2}>Claim link is invalid</Typography.Title>

          <Typography.Text type="secondary">
            In case you think this is an error, please contact the issuer of this claim.
          </Typography.Text>
        </Space>
      );
    }
    return (
      <ErrorResult
        error={shareClaimQRCode.error.message}
        labelRetry="Start again"
        onRetry={onStartAgain}
      />
    );
  }

  return isAsyncTaskStarting(shareClaimQRCode) ? (
    <LoadingResult />
  ) : (
    <Space align="center" direction="vertical" size="large">
      <Avatar
        shape="square"
        size={64}
        src={shareClaimQRCode.data.issuer.logo || IMAGE_PLACEHOLDER_PATH}
      />

      <Space direction="vertical" style={{ textAlign: "center" }}>
        <Typography.Title level={2}>
          {claimQRCheck.status === ClaimQRStatus.Done
            ? "Scan again to add the claim to your wallet"
            : `You received a claim request from ${shareClaimQRCode.data.issuer.displayName}`}
        </Typography.Title>

        <Typography.Text>
          {claimQRCheck.status === ClaimQRStatus.Done
            ? "If you already received a push notification and added the claim to your mobile device, please disregard this message."
            : "Scan the QR code with your Polygon ID wallet to add the claim to your wallet."}
        </Typography.Text>
      </Space>

      {claimQRCheck.status === ClaimQRStatus.Done && (
        <Button icon={<IconRefresh />} onClick={onStartAgain} type="link">
          Start again
        </Button>
      )}

      <Card bodyStyle={{ padding: 0 }} style={{ margin: "auto", width: lg ? 800 : "100%" }}>
        <Row>
          <Col
            md={13}
            style={{
              background:
                'url("/images/noise-bg.png"), linear-gradient(50deg, rgb(130 101 208) 15%, rgba(221, 178, 248, 1) 44%)',
              borderRadius: 8,
              padding: 24,
              width: "100%",
            }}
          >
            <QRCodeSVG
              includeMargin
              level="H"
              style={{ height: "100%", width: "100%" }}
              value={
                claimQRCheck.status === ClaimQRStatus.Done
                  ? JSON.stringify(claimQRCheck.qrcode)
                  : JSON.stringify(shareClaimQRCode.data.qrcode)
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
                {shareClaimQRCode.data.offerDetails.schemaTemplate.schema}
              </Typography.Title>

              <Typography.Title level={5} type="secondary">
                Attributes
              </Typography.Title>

              {shareClaimQRCode.data.offerDetails.attributeValues.map((attribute) => {
                const formattedValue = formatAttributeValue(
                  attribute,
                  shareClaimQRCode.data.offerDetails.schemaTemplate.attributes
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
