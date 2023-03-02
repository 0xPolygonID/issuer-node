import { Button, Card, Col, Divider, Form, Input, Row, Space, Typography, message } from "antd";
import copy from "copy-to-clipboard";
import { generatePath, useNavigate } from "react-router-dom";

import { Claim, claimsQRCreate, claimsQRDownload } from "src/adapters/api/claims";
import { Schema } from "src/adapters/api/schemas";
import { formatAttributeValue } from "src/adapters/parsers/forms";
import { ReactComponent as IconCheckMark } from "src/assets/icons/check.svg";
import { ReactComponent as IconCopy } from "src/assets/icons/copy-01.svg";
import { ReactComponent as ExternalLinkIcon } from "src/assets/icons/link-external-01.svg";
import { ReactComponent as QRIcon } from "src/assets/icons/qr-code.svg";
import { ROUTES } from "src/routes";
import { downloadFile } from "src/utils/browser";
import { CARD_ELLIPSIS_MAXIMUM_WIDTH, FORM_LABEL, SCHEMAS_TABS } from "src/utils/constants";
import { formatDate } from "src/utils/forms";

export function Summary({ claim, schema }: { claim: Claim; schema: Schema }) {
  const navigate = useNavigate();

  const claimLinkURL = `${window.location.origin}${generatePath(ROUTES.claimLink.path, {
    claimID: claim.id,
  })}`;

  const navigateToSharedLinks = () => {
    navigate(
      generatePath(ROUTES.schemas.path, {
        tabID: SCHEMAS_TABS[1].tabID,
      })
    );
  };

  const onCopyToClipboard = () => {
    const hasCopied = copy(claimLinkURL, {
      format: "text/plain",
    });

    if (hasCopied) {
      void message.success("Claim link copied to clipboard.");
    } else {
      void message.error("Couldn't copy claim link. Please try again.");
    }
  };

  const onDownloadQRCode = () => {
    void claimsQRCreate({ id: claim.id }).then((qrData) => {
      if (qrData.isSuccessful) {
        void claimsQRDownload({ claimID: claim.id, sessionID: qrData.data.sessionID }).then(
          (qrBlobData) => {
            if (qrBlobData.isSuccessful) {
              downloadFile(qrBlobData.data, claim.id);
            } else {
              void message.error(qrBlobData.error.message);
            }
          }
        );
      } else {
        void message.error(qrData.error.message);
      }
    });
  };

  return (
    <Card
      className="claiming-card"
      extra={
        <Row>
          <Button icon={<QRIcon />} onClick={onDownloadQRCode} type="link">
            Download QR code
          </Button>

          <Button
            href={generatePath(ROUTES.claimLink.path, { claimID: claim.id })}
            icon={<ExternalLinkIcon />}
            target="_blank"
            type="link"
          >
            Preview claim link
          </Button>
        </Row>
      }
      title={FORM_LABEL.CLAIM_LINK}
    >
      <>
        <Form layout="vertical">
          <Form.Item label="Link">
            <Input.Group className="input-copy-group" compact>
              <Input disabled value={claimLinkURL} />

              <Button icon={<IconCopy style={{ marginRight: 0 }} />} onClick={onCopyToClipboard} />
            </Input.Group>
          </Form.Item>

          <Form.Item label="Details">
            <Card className="bg-light">
              <Space direction="vertical">
                <Row justify="space-between">
                  <Typography.Text type="secondary">{FORM_LABEL.SCHEMA_NAME}</Typography.Text>

                  <Typography.Text
                    ellipsis={{ tooltip: true }}
                    style={{ maxWidth: CARD_ELLIPSIS_MAXIMUM_WIDTH }}
                  >
                    {schema.schema}
                  </Typography.Text>
                </Row>

                {claim.attributeValues.map((attribute, index) => {
                  const formattedValue = formatAttributeValue(attribute, schema.attributes);

                  return (
                    <Row justify="space-between" key={attribute.attributeKey}>
                      <Typography.Text type="secondary">
                        {`Attribute #${index + 1}`}
                      </Typography.Text>

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
                })}

                <Row justify="space-between">
                  <Typography.Text type="secondary">{FORM_LABEL.LINK_VALIDITY}</Typography.Text>

                  <Typography.Text>
                    {claim.claimLinkExpiration ? formatDate(claim.claimLinkExpiration, true) : "-"}
                  </Typography.Text>
                </Row>

                <Row justify="space-between">
                  <Typography.Text type="secondary">Maximum issuance</Typography.Text>

                  <Typography.Text>{claim.limitedClaims || "-"}</Typography.Text>
                </Row>

                <Row justify="space-between">
                  <Typography.Text type="secondary">{FORM_LABEL.CLAIM_EXPIRATION}</Typography.Text>

                  <Typography.Text>
                    {claim.expiresAt ? formatDate(claim.expiresAt) : "-"}
                  </Typography.Text>
                </Row>

                <Row justify="space-between">
                  <Typography.Text type="secondary">Issue date</Typography.Text>

                  <Typography.Text>{formatDate(claim.createdAt)}</Typography.Text>
                </Row>

                <Row justify="space-between">
                  <Typography.Text type="secondary">{FORM_LABEL.SCHEMA_HASH}</Typography.Text>

                  <Typography.Text
                    copyable={{ icon: [<IconCopy key={0} />, <IconCheckMark key={1} />] }}
                    ellipsis
                  >
                    {schema.schemaHash}
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
      </>
    </Card>
  );
}
