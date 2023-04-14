import { Button, Card, Col, Divider, Form, Input, Row, Space, Typography, message } from "antd";
import copy from "copy-to-clipboard";
import { generatePath, useNavigate } from "react-router-dom";

import { credentialsQRCreate, credentialsQRDownload } from "src/adapters/api/credentials";
import { ReactComponent as IconCheckMark } from "src/assets/icons/check.svg";
import { ReactComponent as IconCopy } from "src/assets/icons/copy-01.svg";
import { ReactComponent as ExternalLinkIcon } from "src/assets/icons/link-external-01.svg";
import { ReactComponent as QRIcon } from "src/assets/icons/qr-code.svg";
import { useEnvContext } from "src/contexts/env";
import { Link, Schema } from "src/domain";
import { ROUTES } from "src/routes";
import { downloadFile } from "src/utils/browser";
import {
  ACCESSIBLE_UNTIL,
  CARD_ELLIPSIS_MAXIMUM_WIDTH,
  CREDENTIALS_TABS,
  CREDENTIAL_LINK,
  ISSUE_DATE,
  SCHEMA_HASH,
} from "src/utils/constants";
import { formatDate } from "src/utils/forms";

export function Summary({ link, schema }: { link: Link; schema: Schema }) {
  const env = useEnvContext();
  const navigate = useNavigate();

  const credentialLinkURL = `${window.location.origin}${generatePath(ROUTES.credentialLink.path, {
    credentialID: link.id,
  })}`;

  const navigateToSharedLinks = () => {
    navigate(
      generatePath(ROUTES.credentials.path, {
        tabID: CREDENTIALS_TABS[1].tabID,
      })
    );
  };

  const onCopyToClipboard = () => {
    const hasCopied = copy(credentialLinkURL, {
      format: "text/plain",
    });

    if (hasCopied) {
      void message.success("Credential link copied to clipboard.");
    } else {
      void message.error("Couldn't copy credential link. Please try again.");
    }
  };

  const onDownloadQRCode = () => {
    void credentialsQRCreate({ env, id: link.id }).then((qrData) => {
      if (qrData.isSuccessful) {
        void credentialsQRDownload({
          credentialID: link.id,
          env,
          sessionID: qrData.data.sessionID,
        }).then((qrBlobData) => {
          if (qrBlobData.isSuccessful) {
            downloadFile(qrBlobData.data, link.id);
          } else {
            void message.error(qrBlobData.error.message);
          }
        });
      } else {
        void message.error(qrData.error.message);
      }
    });
  };

  return (
    <Card
      className="issue-credential-card"
      extra={
        <Row>
          <Button icon={<QRIcon />} onClick={onDownloadQRCode} type="link">
            Download QR code
          </Button>

          <Button
            href={generatePath(ROUTES.credentialLink.path, { credentialID: link.id })}
            icon={<ExternalLinkIcon />}
            target="_blank"
            type="link"
          >
            Preview credential link
          </Button>
        </Row>
      }
      title={CREDENTIAL_LINK}
    >
      <Form layout="vertical">
        <Form.Item label="Link">
          <Input.Group className="input-copy-group" compact>
            <Input disabled value={credentialLinkURL} />

            <Button icon={<IconCopy style={{ marginRight: 0 }} />} onClick={onCopyToClipboard} />
          </Input.Group>
        </Form.Item>

        <Form.Item label="Details">
          <Card className="bg-light">
            <Space direction="vertical">
              <Row justify="space-between">
                <Typography.Text type="secondary">Schema name</Typography.Text>

                <Typography.Text
                  ellipsis={{ tooltip: true }}
                  style={{ maxWidth: CARD_ELLIPSIS_MAXIMUM_WIDTH }}
                >
                  {schema.type}
                </Typography.Text>
              </Row>

              {/* {credential.attributes.map((attribute, index) => {
                //TODO Credentials epic
                const formattedValue = formatAttributeValue(attribute, schema.attributes);
                const formattedValue = formatAttributeValue(attribute, []);

                return (
                  <Row justify="space-between" key={attribute.attributeKey}>
                    <Typography.Text type="secondary">{`Attribute #${index + 1}`}</Typography.Text>

                    <Col style={{ maxWidth: CARD_ELLIPSIS_MAXIMUM_WIDTH }}>
                      <Typography.Text ellipsis={{ tooltip: true }}>
                        {attribute.attributeKey}
                      </Typography.Text>

                      {formattedValue.success ? (
                        <Typography.Text>{`: ${formattedValue.data}`}</Typography.Text>
                      ) : (
                        <Typography.Text type="danger">
                          {` (${formattedValue.error})`}
                        </Typography.Text>
                      )}
                    </Col>
                  </Row>
                );
              })} */}

              <Row justify="space-between">
                <Typography.Text type="secondary">{ACCESSIBLE_UNTIL}</Typography.Text>

                <Typography.Text>
                  {link.expiration ? formatDate(link.expiration, true) : "-"}
                </Typography.Text>
              </Row>

              <Row justify="space-between">
                <Typography.Text type="secondary">Maximum issuance</Typography.Text>

                <Typography.Text>{link.maxIssuance || "-"}</Typography.Text>
              </Row>

              <Row justify="space-between">
                <Typography.Text type="secondary">Credential expiration date</Typography.Text>

                {/* <Typography.Text>
                  {credential.expiresAt ? formatDate(credential.expiresAt) : "-"}
                </Typography.Text> */}
              </Row>

              <Row justify="space-between">
                <Typography.Text type="secondary">{ISSUE_DATE}</Typography.Text>

                {/* <Typography.Text>{formatDate(credential.createdAt)}</Typography.Text> */}
              </Row>

              <Row justify="space-between">
                <Typography.Text type="secondary">{SCHEMA_HASH}</Typography.Text>

                <Typography.Text
                  copyable={{ icon: [<IconCopy key={0} />, <IconCheckMark key={1} />] }}
                  ellipsis
                >
                  {schema.hash}
                </Typography.Text>
              </Row>
            </Space>
          </Card>
        </Form.Item>
      </Form>

      <Divider />

      <Row justify="end">
        <Button onClick={navigateToSharedLinks} type="primary">
          Done
        </Button>
      </Row>
    </Card>
  );
}
