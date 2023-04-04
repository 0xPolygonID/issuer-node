import { Button, Row, Space, Typography, message } from "antd";

import { downloadJsonFromUrl } from "src/adapters/json";
import { ReactComponent as IconBack } from "src/assets/icons/arrow-narrow-left.svg";
import { SchemaViewer } from "src/components/schemas/SchemaViewer";
import { Detail } from "src/components/shared/Detail";
import { Json, JsonLdType, Schema } from "src/domain";
import { getBigint, getSchemaHash } from "src/utils/iden3";

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
            copyable={{ enabled: bigint !== undefined }}
            label="BigInt"
            text={bigint || "An error occurred while calculating BigInt"}
          />

          <Detail
            copyable={{ enabled: schemaHash !== undefined }}
            label="Hash"
            text={schemaHash || "An error occurred while calculating Hash"}
          />

          <Detail copyable={{ enabled: true }} label="URL" text={url} />

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
      }
      jsonLdType={jsonLdType}
      rawJsonLdContext={rawJsonLdContext}
      rawJsonSchema={rawJsonSchema}
      schema={schema}
    />
  );
}
