import { Button, Card, Checkbox, Form, Input, Row, Space, Typography, message } from "antd";
import { useState } from "react";
import { generatePath, useNavigate } from "react-router-dom";

import { PayloadSchemaCreate, payloadSchemaCreate, schemasCreate } from "src/adapters/api/schemas";
import { ReactComponent as IconAdd } from "src/assets/icons/plus.svg";
import { SchemaAttributeField } from "src/components/schemas/SchemaAttributeField";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useAuthContext } from "src/hooks/useAuthContext";
import { ROUTES } from "src/routes";
import { processZodError } from "src/utils/adapters";
import {
  SCHEMAS_TABS,
  SCHEMA_FORM_EXTRA_ALPHA_MESSAGE,
  SCHEMA_FORM_HELP_ALPHA_MESSAGE,
  SCHEMA_FORM_HELP_REQUIRED_MESSAGE,
  SCHEMA_KEY_MAX_LENGTH,
} from "src/utils/constants";
import { alphanumericValidator } from "src/utils/forms";

const NAMES_LIST = {
  attributes: "attributes",
  singleChoiceValues: "values",
};

const INITIAL_VALUES: PayloadSchemaCreate = {
  attributes: [
    {
      description: "",
      name: "",
      technicalName: "",
      type: "boolean",
    },
  ],
  mandatoryExpiration: false,
  schema: "",
  technicalName: "",
};

export function CreateSchema() {
  const [creatingSchema, setCreatingSchema] = useState(false);
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const { account, authToken } = useAuthContext();

  const onAttributeTypeChange = (index: number) => {
    form.setFieldValue([NAMES_LIST.attributes, index, NAMES_LIST.singleChoiceValues], undefined);
  };

  const onSubmit = (data: unknown, goToIssueClaimOnSuccess: boolean) => {
    const parser = payloadSchemaCreate.safeParse(data);

    if (parser.success) {
      if (authToken && account?.organization) {
        setCreatingSchema(true);
        void schemasCreate({
          id: account.organization,
          payload: parser.data,
          token: authToken,
        }).then((response) => {
          if (response.isSuccessful) {
            const { id: schemaID } = response.data;

            if (goToIssueClaimOnSuccess) {
              navigate(
                generatePath(ROUTES.issueClaim.path, {
                  schemaID,
                })
              );
            } else {
              navigate(
                generatePath(ROUTES.schemas.path, {
                  tabID: SCHEMAS_TABS[0].tabID,
                })
              );
            }
            void message.success("Claim schema created.");
          } else {
            void message.error(response.error.message);
          }
          setCreatingSchema(false);
        });
      }
    } else {
      processZodError(parser.error).forEach((error) => void message.error(error));
    }
  };

  const onSubmitAndGoToMySchemas = () => {
    form
      .validateFields()
      .then((data: unknown) => {
        onSubmit(data, false);
      })
      .catch(() => ({}));
  };

  const onSubmitAndIssueClaim = (data: unknown) => {
    onSubmit(data, true);
  };

  return (
    <SiderLayoutContent
      backButtonLink={generatePath(ROUTES.schemas.path, {
        tabID: SCHEMAS_TABS[0].tabID,
      })}
      description="Schemas provide an easy way to standardize the claim issuance process."
      showDivider
      title="Create schema"
    >
      <Card className="claiming-card" title="Define schema">
        <Form
          form={form}
          initialValues={INITIAL_VALUES}
          layout="vertical"
          name="createSchema"
          onFinish={onSubmitAndIssueClaim}
          requiredMark={false}
          validateTrigger="onBlur"
        >
          <Form.Item
            extra="User-friendly name that describes the schema. Shown to end-users."
            label="Schema display name*"
            name="schema"
            rules={[{ message: SCHEMA_FORM_HELP_REQUIRED_MESSAGE, required: true }]}
          >
            <Input maxLength={SCHEMA_KEY_MAX_LENGTH} placeholder="e.g. Schema Name" showCount />
          </Form.Item>

          <Form.Item
            extra={SCHEMA_FORM_EXTRA_ALPHA_MESSAGE}
            label="Schema hidden name*"
            name="technicalName"
            rules={[
              {
                message: SCHEMA_FORM_HELP_REQUIRED_MESSAGE,
                required: true,
              },
              {
                message: SCHEMA_FORM_HELP_ALPHA_MESSAGE,
                validator: alphanumericValidator,
              },
            ]}
          >
            <Input maxLength={SCHEMA_KEY_MAX_LENGTH} placeholder="e.g. schemaName" showCount />
          </Form.Item>

          <Form.Item>
            <Form.List name={NAMES_LIST.attributes}>
              {(fields, { add, remove }) => (
                <Space direction="vertical" size="middle">
                  {fields.map(({ key, name }) => (
                    <SchemaAttributeField
                      index={name}
                      key={key}
                      onAttributeTypeChange={onAttributeTypeChange}
                      onRemove={fields.length >= 2 ? remove : undefined}
                      singleChoiceValuesName={NAMES_LIST.singleChoiceValues}
                    />
                  ))}

                  {fields.length < 2 && (
                    <Button
                      icon={<IconAdd />}
                      onClick={() => add({ type: "boolean" })}
                      size="small"
                      type="link"
                    >
                      Add attribute field
                    </Button>
                  )}
                </Space>
              )}
            </Form.List>
          </Form.Item>

          <Form.Item
            extra={
              <Row style={{ marginLeft: 32 }}>
                <Typography.Text>
                  When issuing a claim, there will be a requirement to fill the expiration date.
                </Typography.Text>

                <Typography.Paragraph>
                  Leaving this unchecked will keep the expiration date as optional.
                </Typography.Paragraph>
              </Row>
            }
            name="mandatoryExpiration"
            valuePropName="checked"
          >
            <Checkbox>Mandatory claim expiration date (Optional)</Checkbox>
          </Form.Item>

          <Form.Item style={{ textAlign: "right" }}>
            <Space size="middle">
              <Button htmlType="button" loading={creatingSchema} onClick={onSubmitAndGoToMySchemas}>
                Save schema
              </Button>

              <Button htmlType="submit" loading={creatingSchema} type="primary">
                Save & Issue claim
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </SiderLayoutContent>
  );
}
