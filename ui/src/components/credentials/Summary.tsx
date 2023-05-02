import { Button, Card, Divider, Form, Input, Row, message } from "antd";
import copy from "copy-to-clipboard";
import { generatePath, useNavigate } from "react-router-dom";

import { ReactComponent as IconCopy } from "src/assets/icons/copy-01.svg";
import { ReactComponent as ExternalLinkIcon } from "src/assets/icons/link-external-01.svg";
import { ROUTES } from "src/routes";
import { CREDENTIALS_TABS, CREDENTIAL_LINK } from "src/utils/constants";

export function Summary({ linkID }: { linkID: string }) {
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
      void message.success("Credential link copied to clipboard.");
    } else {
      void message.error("Couldn't copy credential link. Please try again.");
    }
  };

  return (
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
          <Input.Group className="input-copy-group" compact>
            <Input disabled value={linkURL} />

            <Button icon={<IconCopy style={{ marginRight: 0 }} />} onClick={onCopyToClipboard} />
          </Input.Group>
        </Form.Item>
      </Form>

      <Divider />

      <Row justify="end">
        <Button onClick={navigateToLinks} type="primary">
          Done
        </Button>
      </Row>
    </Card>
  );
}
