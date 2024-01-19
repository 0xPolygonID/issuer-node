import { Avatar, Card, Col, Grid, Image, Row, Space, Tabs, TabsProps, Typography } from "antd";
import { QRCodeSVG } from "qrcode.react";
import { ReactNode } from "react";

import { useEnvContext } from "src/contexts/Env";
import { WALLET_APP_STORE_URL, WALLET_PLAY_STORE_URL } from "src/utils/constants";

export function CredentialQR({
  qrCodeBase64,
  qrCodeLink,
  qrCodeRaw,
  schemaType,
  subTitle,
}: {
  qrCodeBase64: string;
  qrCodeLink: string;
  qrCodeRaw: string;
  schemaType: string;
  subTitle: ReactNode;
}) {
  const { issuer } = useEnvContext();

  const { lg } = Grid.useBreakpoint();

  const qrCodeTabs: TabsProps["items"] = [
    {
      children: (
        <QRCodeSVG
          className="full-width"
          includeMargin
          level="H"
          style={{ height: 300 }}
          value={qrCodeLink}
        />
      ),
      key: "1",
      label: "Link",
    },
    {
      children: (
        <QRCodeSVG
          className="full-width"
          includeMargin
          level="H"
          style={{ height: 300 }}
          value={qrCodeRaw}
        />
      ),
      key: "2",
      label: "Raw JSON",
    },
    {
      children: (
        <QRCodeSVG
          className="full-width"
          includeMargin
          level="H"
          style={{ height: 300 }}
          value={qrCodeBase64}
        />
      ),
      key: "3",
      label: "Base64 encoded",
    }
  ];

  return (
    <Space align="center" direction="vertical" size="large">
      <Avatar shape="square" size={64} src={issuer.logo} />

      <Space
        direction="vertical"
        style={{ padding: "0 24px", textAlign: "center", width: lg ? 800 : "100%" }}
      >
        <Typography.Title level={2}>{issuer.name} wants to send you a credential</Typography.Title>

        <Typography.Text style={{ fontSize: 18 }} type="secondary">
          {subTitle}
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
      <Tabs
        className="tab-responsive"
        destroyInactiveTabPane
        items={qrCodeTabs}
        style={{ margin: "auto", width: lg ? 800 : "100%" }}
      />
      <Card bodyStyle={{ padding: 0 }} style={{ margin: "auto", width: lg ? 800 : "100%" }}>
        {schemaType && (
          <Row>
            <Col
              style={{
                padding: 24,
                paddingBottom: 8,
              }}
            >
              <Typography.Title ellipsis={{ tooltip: true }} level={3}>
                {schemaType}
              </Typography.Title>
            </Col>
          </Row>
        )}
      </Card>
    </Space>
  );
}
