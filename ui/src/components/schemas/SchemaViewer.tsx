import { Button, Card, Dropdown, Flex, Form, Row, Select, Space, Typography } from "antd";
import { ReactNode, useState } from "react";
import { UpdateSchema } from "src/adapters/api/schemas";

import ChevronDownIcon from "src/assets/icons/chevron-down.svg?react";
import { JSONHighlighter } from "src/components/schemas/JSONHighlighter";
import { SchemaTree } from "src/components/schemas/SchemaTree";
import { DisplayMethod, Json, JsonLdType, JsonSchema } from "src/domain";

type JsonView = "formatted" | "jsonLdContext" | "jsonSchema";

const JSON_VIEW_LABELS: Record<JsonView, string> = {
  formatted: "Formatted",
  jsonLdContext: "JSON LD Context",
  jsonSchema: "JSON Schema",
};

export function SchemaViewer({
  actions,
  contents,
  displayMethodID,
  displayMethods,
  jsonLdContextObject,
  jsonLdType,
  jsonSchema,
  jsonSchemaObject,
  onEdit,
}: {
  actions: ReactNode;
  contents: ReactNode;
  jsonLdContextObject: Json;
  jsonLdType: JsonLdType;
  jsonSchema: JsonSchema;
  jsonSchemaObject: Json;
} & (
  | { displayMethodID?: never; displayMethods?: undefined; onEdit?: never }
  | {
      displayMethodID: string | null;
      displayMethods: DisplayMethod[];
      onEdit: (formValues: UpdateSchema) => void;
    }
)) {
  const [form] = Form.useForm<UpdateSchema>();
  const [jsonView, setJsonView] = useState<JsonView>("formatted");

  const {
    schema: { description, title },
  } = jsonSchema;
  return (
    <Card className="centered" title={title || jsonLdType.name}>
      <Space direction="vertical" size="large">
        <Card.Meta description={description} />
        <Card className="background-grey">{contents}</Card>

        {displayMethods?.length && (
          <Form
            form={form}
            initialValues={{
              displayMethodID,
            }}
            layout="vertical"
            onValuesChange={(_, formValues: UpdateSchema) => {
              onEdit(formValues);
            }}
          >
            <Form.Item label="Default display method" name="displayMethodID">
              <Select className="full-width" placeholder="Choose the default display method">
                <Select.Option value={null}>None</Select.Option>
                {Object.values(displayMethods).map((displayMethods) => (
                  <Select.Option key={displayMethods.id} value={displayMethods.id}>
                    {displayMethods.name}
                  </Select.Option>
                ))}
              </Select>
            </Form.Item>
          </Form>
        )}

        <Card className="background-grey">
          <Space direction="vertical">
            <Flex align="center" justify="space-between">
              <Typography.Text type="secondary">ATTRIBUTES</Typography.Text>
              <Dropdown
                menu={{
                  items: [
                    {
                      key: "formatted",
                      label: JSON_VIEW_LABELS["formatted"],
                      onClick: () => {
                        setJsonView("formatted");
                      },
                    },
                    {
                      key: "jsonLdContext",
                      label: JSON_VIEW_LABELS["jsonLdContext"],
                      onClick: () => {
                        setJsonView("jsonLdContext");
                      },
                    },
                    {
                      key: "jsonSchema",
                      label: JSON_VIEW_LABELS["jsonSchema"],
                      onClick: () => {
                        setJsonView("jsonSchema");
                      },
                    },
                  ],
                }}
              >
                <Button>
                  {JSON_VIEW_LABELS[jsonView]} <ChevronDownIcon />
                </Button>
              </Dropdown>
            </Flex>
            {(() => {
              switch (jsonView) {
                case "formatted": {
                  return <SchemaTree className="background-grey" jsonSchema={jsonSchema} />;
                }
                case "jsonLdContext": {
                  return <JSONHighlighter json={jsonLdContextObject} />;
                }
                case "jsonSchema": {
                  return <JSONHighlighter json={jsonSchemaObject} />;
                }
              }
            })()}
          </Space>
        </Card>

        <Row justify="end">{actions}</Row>
      </Space>
    </Card>
  );
}
