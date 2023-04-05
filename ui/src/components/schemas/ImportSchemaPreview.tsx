import { Button, Row, Space, Typography, message } from "antd";

import { downloadJsonFromUrl } from "src/adapters/json";
import { ReactComponent as IconBack } from "src/assets/icons/arrow-narrow-left.svg";
import { Detail } from "src/components/schemas/Detail";
import { SchemaViewer } from "src/components/schemas/SchemaViewer";
import { Json, JsonLdType, JsonSchema } from "src/domain";
import { getBigint, getSchemaHash } from "src/utils/iden3";

export function ImportSchemaPreview({
  jsonLdType,
  jsonSchema,
  onBack,
  onImport,
  rawJsonLdContext,
  rawJsonSchema,
  url,
}: {
  jsonLdType: JsonLdType;
  jsonSchema: JsonSchema;
  onBack: () => void;
  onImport: () => void;
  rawJsonLdContext: Json;
  rawJsonSchema: Json;
  url: string;
}) {
  const schemaHashResult = getSchemaHash(jsonLdType);
  const schemaHash =
    schemaHashResult && schemaHashResult.success ? schemaHashResult.data : undefined;

  const bigintResult = getBigint(jsonLdType);
  const bigint = bigintResult && bigintResult.success ? bigintResult.data : undefined;

  return (
    <SchemaViewer
      actions={
        <Space size="middle">
          <Button icon={<IconBack />} onClick={onBack} type="default">
            Previous step
          </Button>

          <Button onClick={onImport} type="primary">
            Import
          </Button>
        </Space>
      }
      contents={
        <Space direction="vertical">
          <Typography.Text type="secondary">SCHEMA DETAILS</Typography.Text>

          <Detail
            copyable={bigint !== undefined}
            data={bigint || "An error occurred while calculating BigInt"}
            label="BigInt"
          />

          <Detail
            copyable={schemaHash !== undefined}
            data={schemaHash || "An error occurred while calculating Hash"}
            label="Hash"
          />

          <Detail copyable data={url} label="URL" />

          <Row justify="space-between">
            <Typography.Text type="secondary">Download</Typography.Text>

            <Button
              onClick={() => {
                downloadJsonFromUrl({ fileName: jsonSchema.name, url })
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
      }
      jsonLdType={jsonLdType}
      jsonSchema={jsonSchema}
      rawJsonLdContext={rawJsonLdContext}
      rawJsonSchema={rawJsonSchema}
    />
  );
}
