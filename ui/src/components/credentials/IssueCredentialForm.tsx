import { Button, DatePicker, Divider, Form, Row, Space, Typography } from "antd";
import dayjs from "dayjs";

import { ReactComponent as IconBack } from "src/assets/icons/arrow-narrow-left.svg";
import { ReactComponent as IconRight } from "src/assets/icons/arrow-narrow-right.svg";
import { ReactComponent as IconCheckMark } from "src/assets/icons/check.svg";
import { ReactComponent as IconCopy } from "src/assets/icons/copy-01.svg";
import { CredentialForm } from "src/components/credentials/CredentialForm";
import { JsonSchema, ObjectAttribute, Schema } from "src/domain";
import { DATE_VALIDITY_MESSAGE, SCHEMA_HASH } from "src/utils/constants";

export type AttributeValues = {
  attributes?: Record<string, unknown>;
  expirationDate?: dayjs.Dayjs;
};

export function IssueCredentialForm({
  initialValues,
  jsonSchema,
  onBack,
  onSubmit,
  schema,
}: {
  initialValues: AttributeValues;
  jsonSchema: JsonSchema;
  onBack: () => void;
  onSubmit: (values: AttributeValues) => void;
  schema: Schema;
}) {
  const credentialSubject =
    (jsonSchema.type === "object" &&
      jsonSchema.schema.properties?.find(
        (child): child is ObjectAttribute => child.type === "object"
      )) ||
    null;

  return (
    credentialSubject && (
      <Form
        initialValues={initialValues}
        layout="vertical"
        name="issueCredentialAttributes"
        onFinish={onSubmit}
        requiredMark={false}
        validateTrigger="onBlur"
      >
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

        <Divider />

        <Typography.Paragraph>{schema.type}</Typography.Paragraph>

        {jsonSchema.type !== "multi" && jsonSchema.schema.description && (
          <Typography.Paragraph type="secondary">
            {jsonSchema.schema.description}
          </Typography.Paragraph>
        )}

        <Space direction="vertical" size="large">
          <CredentialForm attributes={credentialSubject.schema.properties || []} />

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
        </Space>

        <Divider />

        <Row justify="end">
          <Space size="middle">
            <Button icon={<IconBack />} onClick={onBack} type="default">
              Previous step
            </Button>

            <Button disabled={!schema} htmlType="submit" type="primary">
              Next step
              <IconRight />
            </Button>
          </Space>
        </Row>
      </Form>
    )
  );
}
