import Ajv, { ErrorObject } from "ajv";
import addFormats from "ajv-formats";
import { Button, Checkbox, DatePicker, Divider, Form, Row, Space, Typography } from "antd";
import { useState } from "react";

import { IssueCredentialFormData, serializeSchemaForm } from "src/adapters/parsers/forms";
import { ReactComponent as IconBack } from "src/assets/icons/arrow-narrow-left.svg";
import { ReactComponent as IconRight } from "src/assets/icons/arrow-narrow-right.svg";
import { ReactComponent as IconCheckMark } from "src/assets/icons/check.svg";
import { ReactComponent as IconCopy } from "src/assets/icons/copy-01.svg";
import { InputErrors, ObjectAttributeForm } from "src/components/credentials/ObjectAttributeForm";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { JsonSchema, ObjectAttribute, Schema } from "src/domain";
import { ISSUE_CREDENTIAL_DIRECT, ISSUE_CREDENTIAL_LINK, SCHEMA_HASH } from "src/utils/constants";
import { buildAppError, notifyError, notifyParseError } from "src/utils/error";
import { extractCredentialSubjectAttributeWithoutId } from "src/utils/jsonSchemas";

function addErrorToPath(inputErrors: InputErrors, path: string[], error: string): InputErrors {
  const key = path[0];
  if (path.length > 1) {
    const value = (key && inputErrors[key]) || {};
    return key
      ? {
          ...inputErrors,
          [key]: addErrorToPath(
            typeof value === "string" ? {} : value,
            path.slice(1, path.length),
            error
          ),
        }
      : inputErrors;
  } else {
    return key ? { ...inputErrors, [key]: error } : inputErrors;
  }
}

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
  const [inputErrors, setInputErrors] = useState<InputErrors>();

  function isFormValid(value: Record<string, unknown>, objectAttribute: ObjectAttribute): boolean {
    const serializedSchemaForm = serializeSchemaForm({
      attribute: objectAttribute,
      ignoreRequired: true,
      value,
    });
    if (serializedSchemaForm.success) {
      const { properties, required, type } = objectAttribute.schema;
      try {
        const ajv = new Ajv({ allErrors: true });
        addFormats(ajv);
        const validate = ajv.compile({
          properties,
          required,
          type,
        });
        const valid = validate(serializedSchemaForm.data);

        if (valid) {
          setInputErrors(undefined);
          return true;
        } else if (validate.errors) {
          setInputErrors(
            validate.errors.reduce((acc: InputErrors, curr: ErrorObject): InputErrors => {
              if (curr.keyword === "required") {
                // filtering out required errors since we manage these from the antd form
                return acc;
              } else {
                const errorMsg = curr.message
                  ? curr.message.charAt(0).toUpperCase() + curr.message.slice(1)
                  : "Unknown validation error";
                const path = curr.instancePath
                  .split("/")
                  .filter((segment) => segment !== "/" && segment !== "");
                return addErrorToPath(acc, path, errorMsg);
              }
            }, {})
          );
        }
      } catch (error) {
        notifyError(buildAppError(error));
        return false;
      }
    } else {
      notifyParseError(serializedSchemaForm.error);
    }
    return false;
  }

  const credentialSubjectAttributeWithoutId =
    extractCredentialSubjectAttributeWithoutId(jsonSchema);

  return credentialSubjectAttributeWithoutId?.schema.attributes ? (
    <Form
      initialValues={initialValues}
      layout="vertical"
      onFinish={(values: IssueCredentialFormData) => {
        if (
          values.credentialSubject &&
          isFormValid(values.credentialSubject, credentialSubjectAttributeWithoutId)
        ) {
          onSubmit(values);
        }
      }}
      onValuesChange={(_, values: IssueCredentialFormData) => {
        values.credentialSubject &&
          isFormValid(values.credentialSubject, credentialSubjectAttributeWithoutId);
      }}
      requiredMark={false}
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
        <ObjectAttributeForm
          attributes={credentialSubjectAttributeWithoutId.schema.attributes}
          inputErrors={inputErrors}
        />

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
            {type === "directIssue" ? ISSUE_CREDENTIAL_DIRECT : ISSUE_CREDENTIAL_LINK}
            {type === "credentialLink" && <IconRight />}
          </Button>
        </Space>
      </Row>
    </Form>
  ) : (
    <ErrorResult error="An error occurred while getting the credentialSubject attributes of the json schema" />
  );
}
