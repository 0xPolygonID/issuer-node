import { Button, Card, Divider, Flex, Row, Tabs, TabsProps, Typography, theme } from "antd";
import { generatePath, useNavigate } from "react-router-dom";
import { AuthMessage } from "src/adapters/api/credentials";

import DownloadIcon from "src/assets/icons/download-01.svg?react";
import { DownloadQRLink } from "src/components/shared/DownloadQRLink";
import { HighlightLink } from "src/components/shared/HighlightLink";
import { ROUTES } from "src/routes";
import { CREDENTIALS_TABS, CREDENTIAL_LINK } from "src/utils/constants";

function QRTab({
  description,
  fileName,
  link,
  openable,
}: {
  description: string;
  fileName: string;
  link: string;
  openable: boolean;
}) {
  const { token } = theme.useToken();

  return (
    <Flex gap={16} vertical>
      <Typography.Text type="secondary">{description}</Typography.Text>
      <HighlightLink link={link} openable={openable} />
      <Card style={{ alignSelf: "center" }}>
        <DownloadQRLink
          button={
            <Button
              icon={<DownloadIcon />}
              style={{ borderColor: token.colorTextSecondary, color: token.colorTextSecondary }}
            >
              Download QR
            </Button>
          }
          fileName={fileName}
          hidden={false}
          link={link}
        />
      </Card>
    </Flex>
  );
}

export function Summary({ authMessage }: { authMessage: AuthMessage }) {
  const navigate = useNavigate();

  const items: TabsProps["items"] = [
    {
      children: (
        <QRTab
          description="When the recipient interacts with the universal link, it will launch the Privado ID web or mobile wallet interface, displaying the credential offer."
          fileName="Universal link"
          link={authMessage.universalLink}
          openable={true}
        />
      ),
      key: "1",
      label: "Universal link",
    },
    {
      children: (
        <QRTab
          description="When the recipient interacts with the deep link with supported identity wallets, they will receive a credential offer."
          fileName="Deep link"
          link={authMessage.deepLink}
          openable={false}
        />
      ),
      key: "2",
      label: "Deep link",
    },
  ];

  const navigateToLinks = () => {
    navigate(
      generatePath(ROUTES.credentials.path, {
        tabID: CREDENTIALS_TABS[1].tabID,
      })
    );
  };

  return (
    <>
      <Card
        className="issue-credential-card"
        styles={{ body: { paddingTop: 0 }, header: { border: "none" } }}
        title={CREDENTIAL_LINK}
      >
        <Tabs defaultActiveKey="1" items={items} />

        <Divider />

        <Row justify="end">
          <Button onClick={navigateToLinks} type="primary">
            Done
          </Button>
        </Row>
      </Card>
    </>
  );
}
