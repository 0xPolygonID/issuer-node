import { Button, Space, Typography } from "antd";

import IconBack from "src/assets/icons/arrow-narrow-left.svg?react";
import { DownloadSchema } from "src/components/schemas/DownloadSchema";
import { SchemaViewer } from "src/components/schemas/SchemaViewer";
import { Detail } from "src/components/shared/Detail";
import { useEnvContext } from "src/contexts/Env";
import { Json, JsonLdType, JsonSchema } from "src/domain";
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
  jsonLdContextObject: Json;
  jsonLdType: JsonLdType;
  jsonSchema: JsonSchema;
  jsonSchemaObject: Json;
  onBack: () => void;
  onImport: () => void;
  url: string;
}) {
  const env = useEnvContext();
  const bigintResult = getBigint(jsonLdType);
  const bigint = bigintResult && bigintResult.success ? bigintResult.data : null;
  const schemaHashResult = getSchemaHash(jsonLdType);
  const schemaHash = schemaHashResult.success ? schemaHashResult.data : null;
  const version = jsonSchema.jsonSchemaProps.$metadata.version;

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

          <Detail copyable label="Schema type" text={jsonLdType.name} />

          {version && <Detail label="Schema version" text={version} />}

          <Detail
            copyable={bigint !== null}
            label="BigInt"
            text={bigint || "An error occurred while calculating BigInt."}
          />

          <Detail
            copyable={schemaHash !== null}
            label="Hash"
            text={schemaHash || "An error occurred while calculating Hash."}
          />

          <Detail copyable href={url} label="URL" text={url} />

          <DownloadSchema env={env} fileName={jsonSchema.name} url={url} />
        </Space>
      }
      jsonLdContextObject={jsonLdContextObject}
      jsonLdType={jsonLdType}
      jsonSchema={jsonSchema}
      jsonSchemaObject={jsonSchemaObject}
    />
  );
}
