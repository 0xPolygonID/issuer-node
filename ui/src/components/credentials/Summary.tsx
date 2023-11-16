import { Button, Card, Divider, Form, Input, Row, Space, message } from "antd";
import copy from "copy-to-clipboard";
import { generatePath, useNavigate } from "react-router-dom";

import IconCopy from "src/assets/icons/copy-01.svg?react";
import ExternalLinkIcon from "src/assets/icons/link-external-01.svg?react";
import { ROUTES } from "src/routes";
import { CREDENTIALS_TABS, CREDENTIAL_LINK } from "src/utils/constants";

export function Summary({ linkID }: { linkID: string }) {
  const [messageAPI, messageContext] = message.useMessage();
  const navigate = useNavigate();

  const linkURL = `${window.location.origin}${generatePath(ROUTES.credentialLinkQR.path, {
    linkID,
  })}`;

  const navigateToLinks = () => {
    navigate(
      generatePath(ROUTES.credentials.path, {
        tabID: CREDENTIALS_TABS[1].tabID,
      })
    );
  };

  const onCopyToClipboard = () => {
    const hasCopied = copy(linkURL, {
      format: "text/plain",
    });

    if (hasCopied) {
      void messageAPI.success("Credential link copied to clipboard.");
    } else {
      void messageAPI.error("Couldn't copy credential link. Please try again.");
    }
  };

  return (
    <>
      {messageContext}

      <Card
        className="issue-credential-card"
        extra={
          <Button
            href={generatePath(ROUTES.credentialLinkQR.path, { linkID })}
            icon={<ExternalLinkIcon />}
            target="_blank"
            type="link"
          >
            View link
          </Button>
        }
        title={CREDENTIAL_LINK}
      >
        <Form layout="vertical">
          <Form.Item>
            <Space.Compact className="full-width">
              <Input allowClear disabled value={linkURL} />

              <Button icon={<IconCopy style={{ marginRight: 0 }} />} onClick={onCopyToClipboard} />
            </Space.Compact>
          </Form.Item>
        </Form>

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
