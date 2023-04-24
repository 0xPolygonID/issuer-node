import { Button, Checkbox, DatePicker, Divider, Form, Row, Space, Typography } from "antd";
import dayjs from "dayjs";
import { IssueCredentialFormData } from "src/adapters/parsers/forms";

import { ReactComponent as IconBack } from "src/assets/icons/arrow-narrow-left.svg";
import { ReactComponent as IconRight } from "src/assets/icons/arrow-narrow-right.svg";
import { ReactComponent as IconCheckMark } from "src/assets/icons/check.svg";
import { ReactComponent as IconCopy } from "src/assets/icons/copy-01.svg";
import { CredentialSubjectForm } from "src/components/credentials/CredentialSubjectForm";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { JsonSchema, ObjectAttribute, Schema } from "src/domain";
import { DATE_VALIDITY_MESSAGE, SCHEMA_HASH } from "src/utils/constants";

export function IssueCredentialForm({
  initialValues,
  jsonSchema,
  onBack,
  onSubmit,
  schema,
  type,
}: {
  initialValues: IssueCredentialFormData;
  jsonSchema: JsonSchema;
  onBack: () => void;
  onSubmit: (values: IssueCredentialFormData) => void;
  schema: Schema;
  type: "directIssue" | "credentialLink";
}) {
  const rawCredentialSubjectAttributes =
    (jsonSchema.type === "object" &&
      jsonSchema.schema.properties
        ?.filter((child): child is ObjectAttribute => child.type === "object")
        .find((child) => child.name === "credentialSubject")?.schema.properties) ||
    null;

  // When creating a link, we are unable to provide the identity holder's did on the credential issuance form
  // because the connection may not yet exist. Therefore, we have to filter this field and skip its validation.
  const credentialSubjectAttributes =
    type === "directIssue"
      ? rawCredentialSubjectAttributes
      : rawCredentialSubjectAttributes &&
        rawCredentialSubjectAttributes.filter((attribute) => attribute.name !== "id");

  return credentialSubjectAttributes ? (
    <Form
      initialValues={initialValues}
      layout="vertical"
      onFinish={onSubmit}
      requiredMark={false}
      validateTrigger="onBlur"
    >
      <Form.Item>
        <Space direction="vertical">
          <Row justify="space-between">
            <Typography.Text type="secondary">{SCHEMA_HASH}</Typography.Text>

            <Typography.Text copyable={{ icon: [<IconCopy key={0} />, <IconCheckMark key={1} />] }}>
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
        <CredentialSubjectForm attributes={credentialSubjectAttributes} />

        <Form.Item
          label="Credential expiration date"
          name="expirationDate"
          rules={[{ message: DATE_VALIDITY_MESSAGE, required: false }]}
        >
          <DatePicker disabledDate={(current) => current < dayjs()} />
        </Form.Item>
      </Space>

      <Form.Item
        label="Proof type"
        name="proofTypes"
        rules={[{ message: "At least one proof type is required.", required: true }]}
      >
        <Checkbox.Group>
          <Form.Item>
            <Checkbox value="SIG">
              <Typography.Text>Signature-based (SIG)</Typography.Text>

              <Typography.Text type="secondary">
                Credential signed by the issuer using a BJJ private key.
              </Typography.Text>
            </Checkbox>
          </Form.Item>

          <Form.Item>
            <Checkbox value="MTP">
              <Typography.Text>Merkle Tree Proof (MTP)</Typography.Text>

              <Typography.Text type="secondary">
                Credential will be added to the issuer&apos;s state tree. The state transition
                involves an on-chain transaction and gas fees.
              </Typography.Text>
            </Checkbox>
          </Form.Item>
        </Checkbox.Group>
      </Form.Item>

      <Divider />

      <Row justify="end">
        <Space size="middle">
          <Button icon={<IconBack />} onClick={onBack} type="default">
            Previous step
          </Button>

          <Button disabled={!schema} htmlType="submit" type="primary">
            Create credential link
            <IconRight />
          </Button>
        </Space>
      </Row>
    </Form>
  ) : (
    <ErrorResult error="An error occurred while getting the credentialSubject attributes of the json schema" />
  );
}
