import { Button, DatePicker, Divider, Form, Row, Space, Typography } from "antd";
import dayjs from "dayjs";

import { SchemaPayload } from "src/adapters/api/schemas";
import { ReactComponent as IconRight } from "src/assets/icons/arrow-narrow-right.svg";
import { ReactComponent as IconCheckMark } from "src/assets/icons/check.svg";
import { ReactComponent as IconCopy } from "src/assets/icons/copy-01.svg";
import { CredentialAttribute } from "src/components/credentials/CredentialAttribute";
import { DATE_VALIDITY_MESSAGE, SCHEMA_HASH } from "src/utils/constants";

export type AttributeValues = {
  attributes?: Record<string, unknown>;
  expirationDate?: dayjs.Dayjs;
};

export function IssueCredentialForm({
  initialValues,
  onSubmit,
  schema,
}: {
  initialValues: AttributeValues;
  onSubmit: (values: AttributeValues) => void;
  schema?: SchemaPayload;
}) {
  return (
    <Form
      initialValues={initialValues}
      layout="vertical"
      name="issueCredentialAttributes"
      onFinish={onSubmit}
      requiredMark={false}
      validateTrigger="onBlur"
    >
      {schema && (
        <>
          <Form.Item>
            <Space direction="vertical">
              <Row justify="space-between">
                <Typography.Text type="secondary">{SCHEMA_HASH}</Typography.Text>

                <Typography.Text
                  copyable={{ icon: [<IconCopy key={0} />, <IconCheckMark key={1} />] }}
                >
                  {schema.hash}
                </Typography.Text>
              </Row>
            </Space>
          </Form.Item>

          <Form.Item>
            <Space direction="vertical" size="middle">
              {/* //TODO Credentials epic */}
              {[].map((schemaAttribute, index) => (
                <CredentialAttribute index={index} key={index} schemaAttribute={schemaAttribute} />
              ))}
            </Space>
          </Form.Item>

          <Form.Item
            label="Credential expiration date"
            name="expirationDate"
            rules={[{ message: DATE_VALIDITY_MESSAGE, required: true }]}
            // TODO Credentials epic
            // rules={
            //   schema.mandatoryExpiration ? [{ message: DATE_VALIDITY_MESSAGE, required: true }] : []
            // }
          >
            <DatePicker disabledDate={(current) => current < dayjs()} />
          </Form.Item>
        </>
      )}

      <Divider />

      <Row justify="end">
        <Button disabled={!schema} htmlType="submit" type="primary">
          Next step
          <IconRight />
        </Button>
      </Row>
    </Form>
  );
}
