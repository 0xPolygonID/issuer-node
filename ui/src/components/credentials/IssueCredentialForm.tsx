import { Button, Checkbox, DatePicker, Divider, Form, Row, Space, Typography } from "antd";
import { IssueCredentialFormData } from "src/adapters/parsers/forms";

import { ReactComponent as IconBack } from "src/assets/icons/arrow-narrow-left.svg";
import { ReactComponent as IconRight } from "src/assets/icons/arrow-narrow-right.svg";
import { ReactComponent as IconCheckMark } from "src/assets/icons/check.svg";
import { ReactComponent as IconCopy } from "src/assets/icons/copy-01.svg";
import { CredentialSubjectForm } from "src/components/credentials/CredentialSubjectForm";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { JsonSchema, ObjectAttribute, Schema } from "src/domain";
import { SCHEMA_HASH } from "src/utils/constants";

export function IssueCredentialForm({
  initialValues,
  jsonSchema,
  loading,
  onBack,
  onSubmit,
  schema,
  type,
}: {
  initialValues: IssueCredentialFormData;
  jsonSchema: JsonSchema;
  loading: boolean;
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

  const credentialSubjectAttributes =
    rawCredentialSubjectAttributes &&
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

        <Form.Item label="Proof type" name="proofTypes" required>
          <Checkbox.Group>
            <Space direction="vertical">
              <Checkbox value="SIG">
                <Typography.Text>Signature-based (SIG)</Typography.Text>

                <Typography.Text type="secondary">
                  Credential signed by the issuer using a BJJ private key.
                </Typography.Text>
              </Checkbox>

              <Checkbox value="MTP">
                <Typography.Text>Merkle Tree Proof (MTP)</Typography.Text>

                <Typography.Text type="secondary">
                  Credential will be added to the issuer&apos;s state tree. The state transition
                  involves an on-chain transaction and gas fees.
                </Typography.Text>
              </Checkbox>
            </Space>
          </Checkbox.Group>
        </Form.Item>
      </Space>

      <Form.Item label="Credential expiration date" name="credentialExpiration">
        <DatePicker />
      </Form.Item>

      <Divider />

      <Row justify="end">
        <Space size="middle">
          <Button icon={<IconBack />} onClick={onBack} type="default">
            Previous step
          </Button>

          <Button disabled={!schema} htmlType="submit" loading={loading} type="primary">
            {type === "directIssue" ? "Issue credential directly" : "Create credential link"}
            {type === "credentialLink" && <IconRight />}
          </Button>
        </Space>
      </Row>
    </Form>
  ) : (
    <ErrorResult error="An error occurred while getting the credentialSubject attributes of the json schema" />
  );
}
