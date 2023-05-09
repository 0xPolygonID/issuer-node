import { Button, Row, Space, Typography, message } from "antd";

import { downloadJsonFromUrl } from "src/adapters/json";
import { ReactComponent as IconBack } from "src/assets/icons/arrow-narrow-left.svg";
import { SchemaViewer } from "src/components/schemas/SchemaViewer";
import { Detail } from "src/components/shared/Detail";
import { JsonLdType, JsonSchema } from "src/domain";
import { getBigint, getSchemaHash } from "src/utils/iden3";

export function ImportSchemaPreview({
  jsonLdContextObject,
  jsonLdType,
  jsonSchema,
  jsonSchemaObject,
  onBack,
  onImport,
  url,
}: {
  jsonLdContextObject: Record<string, unknown>;
  jsonLdType: JsonLdType;
  jsonSchema: JsonSchema;
  jsonSchemaObject: Record<string, unknown>;
  onBack: () => void;
  onImport: () => void;
  url: string;
}) {
  const bigintResult = getBigint(jsonLdType);
  const bigint = bigintResult && bigintResult.success ? bigintResult.data : null;
  const schemaHashResult = getSchemaHash(jsonLdType);
  const schemaHash = schemaHashResult && schemaHashResult.success ? schemaHashResult.data : null;

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
            copyable={bigint !== null}
            label="BigInt"
            text={bigint || "An error occurred while calculating BigInt"}
          />

          <Detail
            copyable={schemaHash !== null}
            label="Hash"
            text={schemaHash || "An error occurred while calculating Hash"}
          />

          <Detail copyable label="URL" text={url} />

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
      jsonLdContextObject={jsonLdContextObject}
      jsonLdType={jsonLdType}
      jsonSchema={jsonSchema}
      jsonSchemaObject={jsonSchemaObject}
    />
  );
}
