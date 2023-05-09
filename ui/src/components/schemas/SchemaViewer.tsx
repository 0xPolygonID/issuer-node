import { Button, Card, Dropdown, Row, Space, Typography } from "antd";
import { ReactNode, useState } from "react";

import { ReactComponent as ChevronDownIcon } from "src/assets/icons/chevron-down.svg";
import { JSONHighlighter } from "src/components/schemas/JSONHighlighter";
import { SchemaTree } from "src/components/schemas/SchemaTree";
import { JsonLdType, JsonSchema } from "src/domain";

type JsonView = "formatted" | "jsonLdContext" | "jsonSchema";

const JSON_VIEW_LABELS: Record<JsonView, string> = {
  formatted: "Formatted",
  jsonLdContext: "JSON LD Context",
  jsonSchema: "JSON Schema",
};

export function SchemaViewer({
  actions,
  contents,
  jsonLdContextObject,
  jsonLdType,
  jsonSchema,
  jsonSchemaObject,
}: {
  actions: ReactNode;
  contents: ReactNode;
  jsonLdContextObject: Record<string, unknown>;
  jsonLdType: JsonLdType;
  jsonSchema: JsonSchema;
  jsonSchemaObject: Record<string, unknown>;
}) {
  const [jsonView, setJsonView] = useState<JsonView>("formatted");

  return (
    <Card
      className="centered"
      extra={
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
          <Button style={{ margin: "16px 0" }}>
            {JSON_VIEW_LABELS[jsonView]} <ChevronDownIcon />
          </Button>
        </Dropdown>
      }
      title={jsonLdType.name}
    >
      <Space direction="vertical" size="large">
        <Card className="background-grey">{contents}</Card>

        {(() => {
          switch (jsonView) {
            case "formatted": {
              return (
                <Card className="background-grey">
                  <Space direction="vertical">
                    <Typography.Text type="secondary">ATTRIBUTES</Typography.Text>

                    <SchemaTree className="background-grey" jsonSchema={jsonSchema} />
                  </Space>
                </Card>
              );
            }
            case "jsonLdContext": {
              return <JSONHighlighter json={jsonLdContextObject} />;
            }
            case "jsonSchema": {
              return <JSONHighlighter json={jsonSchemaObject} />;
            }
          }
        })()}

        <Row justify="end">{actions}</Row>
      </Space>
    </Card>
  );
}
