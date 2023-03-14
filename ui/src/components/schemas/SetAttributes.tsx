import { Button, Card, DatePicker, Divider, Form, Row, Space, Typography } from "antd";
import dayjs from "dayjs";
import { Schema } from "src/adapters/api/schemas";

import { ReactComponent as IconRight } from "src/assets/icons/arrow-narrow-right.svg";
import { ReactComponent as IconCheckMark } from "src/assets/icons/check.svg";
import { ReactComponent as IconCopy } from "src/assets/icons/copy-01.svg";
import { ClaimAttribute } from "src/components/schemas/ClaimAttribute";
import {
  CARD_ELLIPSIS_MAXIMUM_WIDTH,
  DATE_VALIDITY_MESSAGE,
  FORM_LABEL,
} from "src/utils/constants";

export type AttributeValues = {
  attributes?: Record<string, unknown>;
  expirationDate?: dayjs.Dayjs;
};

export function SetAttributes({
  initialValues,
  onSubmit,
  schema,
}: {
  initialValues: AttributeValues;
  onSubmit: (values: AttributeValues) => void;
  schema: Schema;
}) {
  return (
    <Card className="claiming-card" title="Schema">
      <Form
        initialValues={initialValues}
        layout="vertical"
        name="issueClaimAttributes"
        onFinish={onSubmit}
        requiredMark={false}
        validateTrigger="onBlur"
      >
        <Form.Item>
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

            <Row justify="space-between">
              <Typography.Text type="secondary">{FORM_LABEL.SCHEMA_HASH}</Typography.Text>

              <Typography.Text
                copyable={{ icon: [<IconCopy key={0} />, <IconCheckMark key={1} />] }}
              >
                {schema.schemaHash}
              </Typography.Text>
            </Row>
          </Space>
        </Form.Item>

        <Form.Item>
          <Space direction="vertical" size="middle">
            {schema.attributes.map((schemaAttribute, index) => (
              <ClaimAttribute index={index} key={index} schemaAttribute={schemaAttribute} />
            ))}
          </Space>
        </Form.Item>

        <Form.Item
          label={`${FORM_LABEL.CLAIM_EXPIRATION}${schema.mandatoryExpiration ? "*" : ""}`}
          name="expirationDate"
          rules={
            schema.mandatoryExpiration ? [{ message: DATE_VALIDITY_MESSAGE, required: true }] : []
          }
        >
          <DatePicker disabledDate={(current) => current < dayjs()} />
        </Form.Item>

        <Divider />

        <Row justify="end">
          <Button htmlType="submit" type="primary">
            Next step
            <IconRight />
          </Button>
        </Row>
      </Form>
    </Card>
  );
}
