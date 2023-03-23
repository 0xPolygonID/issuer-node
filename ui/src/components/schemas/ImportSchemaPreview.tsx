import { Button, Card, Col, Dropdown, Row, Space, Typography, message } from "antd";
import { useState } from "react";
import SyntaxHighlighter from "react-syntax-highlighter";
import { a11yLight } from "react-syntax-highlighter/dist/esm/styles/hljs";

import { ReactComponent as IconBack } from "src/assets/icons/arrow-narrow-left.svg";
import { ReactComponent as ChevronDownIcon } from "src/assets/icons/chevron-down.svg";
import { Detail } from "src/components/schemas/Detail";
import { SchemaTree } from "src/components/schemas/SchemaTree";
import { Json, JsonLdType, Schema } from "src/domain";
import { downloadJsonFromUrl } from "src/utils/adapters";
import { CARD_WIDTH } from "src/utils/constants";
import { getBigint, getSchemaHash } from "src/utils/iden3";

type JsonView = "formatted" | "jsonLdContext" | "jsonSchema";

const JSON_VIEW_LABELS: Record<JsonView, string> = {
  formatted: "Formatted",
  jsonLdContext: "JSON LD Context",
  jsonSchema: "JSON Schema",
};

export function ImportSchemaPreview({
  jsonLdType,
  onBack,
  onImport,
  rawJsonLdContext,
  rawJsonSchema,
  schema,
  url,
}: {
  jsonLdType: JsonLdType;
  onBack: () => void;
  onImport: () => void;
  rawJsonLdContext: Json;
  rawJsonSchema: Json;
  schema: Schema;
  url: string;
}) {
  const [jsonView, setJsonView] = useState<JsonView>("formatted");

  const schemaHashResult = getSchemaHash(jsonLdType);
  const schemaHash =
    schemaHashResult && schemaHashResult.success ? schemaHashResult.data : undefined;

  const bigintResult = getBigint(jsonLdType);
  const bigint = bigintResult && bigintResult.success ? bigintResult.data : undefined;

  const DETAIL_WIDTH = 350;

  return (
    <Card
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
      style={{ margin: "auto", maxWidth: CARD_WIDTH }}
      title={jsonLdType.name}
    >
      <Space direction="vertical" size="large">
        <Col style={{ background: "#F2F4F7", borderRadius: 8, padding: 24 }}>
          <Space direction="vertical">
            <Row justify="space-between">
              <Typography.Text type="secondary">SCHEMA DETAILS</Typography.Text>
            </Row>

            <Detail
              copyable={bigint !== undefined}
              data={bigint || "An error occurred while calculating BigInt"}
              label="BigInt"
              maxWidth={DETAIL_WIDTH}
            />

            <Detail
              copyable={schemaHash !== undefined}
              data={schemaHash || "An error occurred while calculating Hash"}
              label="Hash"
              maxWidth={DETAIL_WIDTH}
            />

            <Detail copyable data={url} label="URL" maxWidth={DETAIL_WIDTH} />

            <Row justify="space-between">
              <Typography.Text type="secondary">Download</Typography.Text>

              <Button
                onClick={() => {
                  downloadJsonFromUrl({ fileName: schema.name, url })
                    .then(() => {
                      void message.success("Schema successfully downloaded");
                    })
                    .catch(() => {
                      void message.error(
                        "An error occurred while downloading the schema. Please try again"
                      );
                    });
                }}
                style={{ height: 24, padding: 0 }}
                type="link"
              >
                JSON Schema
              </Button>
            </Row>
          </Space>
        </Col>

        {(() => {
          switch (jsonView) {
            case "formatted": {
              return (
                <Col style={{ background: "#F2F4F7", borderRadius: 8, padding: 24 }}>
                  <Space direction="vertical">
                    <Row justify="space-between">
                      <Typography.Text type="secondary">ATTRIBUTES</Typography.Text>
                    </Row>
                    <SchemaTree schema={schema} style={{ background: "#F2F4F7" }} />
                  </Space>
                </Col>
              );
            }
            case "jsonLdContext": {
              return (
                <SyntaxHighlighter
                  customStyle={{
                    background: "#F2F4F7",
                    borderRadius: 8,
                    margin: 0,
                    padding: 16,
                  }}
                  language="json"
                  style={a11yLight}
                >
                  {JSON.stringify(rawJsonLdContext, null, 2)}
                </SyntaxHighlighter>
              );
            }
            case "jsonSchema": {
              return (
                <SyntaxHighlighter
                  customStyle={{
                    background: "#F2F4F7",
                    borderRadius: 8,
                    margin: 0,
                    padding: 16,
                  }}
                  language="json"
                  style={a11yLight}
                >
                  {JSON.stringify(rawJsonSchema, null, 2)}
                </SyntaxHighlighter>
              );
            }
          }
        })()}

        <Row justify="end">
          <Space size="middle">
            <Button icon={<IconBack />} onClick={onBack} type="default">
              Previous step
            </Button>

            <Button onClick={onImport} type="primary">
              Import
            </Button>
          </Space>
        </Row>
      </Space>
    </Card>
  );
}
