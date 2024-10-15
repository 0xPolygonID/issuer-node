import {
  Button,
  Card,
  Divider,
  Flex,
  Row,
  Tabs,
  TabsProps,
  Typography,
  message,
  theme,
} from "antd";
import { useEffect, useState } from "react";
import { generatePath, useNavigate } from "react-router-dom";

import { AuthRequestMessage, createAuthRequestMessage } from "src/adapters/api/credentials";
import DownloadIcon from "src/assets/icons/download-01.svg?react";
import { DownloadQRLink } from "src/components/shared/DownloadQRLink";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { HighlightLink } from "src/components/shared/HighlightLink";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { AppError } from "src/domain";
import { ROUTES } from "src/routes";
import { AsyncTask, hasAsyncTaskFailed, isAsyncTaskStarting } from "src/utils/async";
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

export function Summary({ linkID }: { linkID: string }) {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();
  const navigate = useNavigate();
  const [authMessage, setAuthMessage] = useState<AsyncTask<AuthRequestMessage, AppError>>({
    status: "pending",
  });
  const [messageAPI, messageContext] = message.useMessage();

  useEffect(() => {
    void createAuthRequestMessage({
      env,
      identifier,
      linkID,
    }).then((response) => {
      if (response.success) {
        setAuthMessage({ data: response.data, status: "successful" });
      } else {
        setAuthMessage({ error: response.error, status: "failed" });
        void messageAPI.error(response.error.message);
      }
    });
  }, [linkID, env, identifier, messageAPI]);

  const navigateToLinks = () => {
    navigate(
      generatePath(ROUTES.credentials.path, {
        tabID: CREDENTIALS_TABS[1].tabID,
      })
    );
  };

  return (
    <>
      {messageContext}

      {(() => {
        if (hasAsyncTaskFailed(authMessage)) {
          return (
            <Card className="centered">
              <ErrorResult error={authMessage.error.message} />;
            </Card>
          );
        } else if (isAsyncTaskStarting(authMessage)) {
          return (
            <Card className="centered">
              <LoadingResult />
            </Card>
          );
        } else {
          const items: TabsProps["items"] = [
            {
              children: (
                <QRTab
                  description="When the recipient interacts with the universal link, it will launch the Privado ID web or mobile wallet interface, displaying the credential offer."
                  fileName="Universal link"
                  link={authMessage.data.universalLink}
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
                  link={authMessage.data.deepLink}
                  openable={false}
                />
              ),
              key: "2",
              label: "Deep link",
            },
          ];

          return (
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
          );
        }
      })()}
    </>
  );
}
