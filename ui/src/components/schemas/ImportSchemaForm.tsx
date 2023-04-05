import { Button, Card, Divider, Form, Input, Radio, Row, Space } from "antd";
import { useState } from "react";
import { z } from "zod";

import { getSchemaFromUrl, getSchemaJsonLdTypes } from "src/adapters/schemas";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { Json, JsonLdType } from "src/domain";
import { JsonSchema } from "src/domain/jsonSchema";
import { AsyncTask, isAsyncTaskDataAvailable } from "src/utils/async";
import { processZodError } from "src/utils/error";

export type FormData = {
  jsonLdType: JsonLdType;
  jsonLdTypes: AsyncTask<JsonLdType[], string | z.ZodError>;
  jsonSchema: JsonSchema;
  rawJsonLdContext: Json;
  rawJsonSchema: Json;
  schemaUrl: string;
  schemaUrlInput: string;
};

export function ImportSchemaForm({
  initialFormData,
  onFinish,
}: {
  initialFormData?: FormData;
  onFinish: (formData: FormData) => void;
}) {
  const [schemaUrlInput, setSchemaUrlInput] = useState<string>(
    initialFormData?.schemaUrlInput || ""
  );
  const [schemaUrl, setSchemaUrl] = useState<string | undefined>(initialFormData?.schemaUrl);
  const [jsonSchema, setJsonSchema] = useState<AsyncTask<JsonSchema, string | z.ZodError>>(
    initialFormData
      ? { data: initialFormData.jsonSchema, status: "successful" }
      : {
          status: "pending",
        }
  );
  const [rawJsonSchema, setRawJsonSchema] = useState<Json | undefined>(
    initialFormData?.rawJsonSchema
  );
  const [jsonLdTypes, setJsonLdTypes] = useState<AsyncTask<JsonLdType[], string | z.ZodError>>(
    initialFormData
      ? initialFormData.jsonLdTypes
      : {
          status: "pending",
        }
  );
  const [jsonLdTypeInput, setJsonLdTypeInput] = useState<JsonLdType | undefined>(
    initialFormData?.jsonLdType
  );
  const [rawJsonLdContext, setRawJsonLdContext] = useState<Json | undefined>(
    initialFormData?.rawJsonLdContext
  );

  const processError = (error: unknown) =>
    error instanceof z.ZodError ? error : error instanceof Error ? error.message : "Unknown error";

  const fetchSchemaFromUrl = (url: string): void => {
    setJsonSchema({ status: "loading" });

    getSchemaFromUrl({
      url,
    })
      .then(([jsonSchema, rawSchema]) => {
        setSchemaUrl(url);
        setJsonSchema({ data: jsonSchema, status: "successful" });
        setRawJsonSchema(rawSchema);
        setJsonLdTypes({ status: "loading" });

        getSchemaJsonLdTypes({
          jsonSchema,
        })
          .then(([jsonLdTypes, rawJsonLdContext]) => {
            setJsonLdTypes({ data: jsonLdTypes, status: "successful" });
            setRawJsonLdContext(rawJsonLdContext);

            if (jsonLdTypes.length === 1) {
              setJsonLdTypeInput(jsonLdTypes[0]);
            }
          })
          .catch((error) => {
            setJsonLdTypes({
              error: processError(error),
              status: "failed",
            });
          });
      })
      .catch((error) => {
        setJsonSchema({
          error: processError(error),
          status: "failed",
        });
      });
  };

  const loadSchema = () => {
    const parsedUrl = z.string().safeParse(schemaUrlInput);

    if (parsedUrl.success) {
      setJsonSchema({ status: "pending" });
      setJsonLdTypes({ status: "pending" });
      setJsonLdTypeInput(undefined);
      fetchSchemaFromUrl(parsedUrl.data);
    } else {
      setJsonSchema({ error: `"${schemaUrlInput}" is not a valid URL`, status: "failed" });
    }
  };

  return (
    <Card className="centered">
      <Space direction="vertical" size="large">
        <Card.Meta
          description="The schema URL must remain publicly accessible after import because the schema will continue to be retrieved in the future."
          title="Provide JSON schema URL"
        />

        <Form
          layout="vertical"
          onFinish={() => {
            if (
              schemaUrl &&
              isAsyncTaskDataAvailable(jsonSchema) &&
              jsonLdTypeInput &&
              rawJsonSchema &&
              rawJsonLdContext
            ) {
              onFinish({
                jsonLdType: jsonLdTypeInput,
                jsonLdTypes: jsonLdTypes,
                jsonSchema: jsonSchema.data,
                rawJsonLdContext,
                rawJsonSchema,
                schemaUrl: schemaUrl,
                schemaUrlInput: schemaUrlInput,
              });
            }
          }}
        >
          <Form.Item label="URL to JSON schema" required>
            <Row>
              <Input
                onChange={(event) => setSchemaUrlInput(event.target.value)}
                onPressEnter={loadSchema}
                placeholder="Enter URL"
                style={{ flex: 1, marginRight: 8 }}
                value={schemaUrlInput}
              />

              <Button onClick={loadSchema}>Fetch</Button>
            </Row>
          </Form.Item>

          {isAsyncTaskDataAvailable(jsonLdTypes) && (
            <Form.Item label="Select schema type" required>
              <Radio.Group value={jsonLdTypeInput?.name}>
                <Space direction="vertical">
                  {jsonLdTypes.data.map((jsonLdType) => (
                    <Radio
                      key={jsonLdType.name}
                      onClick={() => {
                        setJsonLdTypeInput(jsonLdType);
                      }}
                      value={jsonLdType.name}
                    >
                      {jsonLdType.name}
                    </Radio>
                  ))}
                </Space>
              </Radio.Group>
            </Form.Item>
          )}

          {(jsonSchema.status === "loading" || jsonLdTypes.status === "loading") && (
            <LoadingResult />
          )}

          {jsonSchema.status === "failed" && (
            <ErrorResult
              error={
                jsonSchema.error instanceof z.ZodError
                  ? [
                      "An error occurred while trying to parse this schema:",
                      ...processZodError(jsonSchema.error).map((e) => `"${e}"`),
                      "Please provide a valid JSON Schema.",
                    ].join("\n")
                  : `An error occurred while downloading this schema:\n"${jsonSchema.error}"\nPlease try again.`
              }
            />
          )}

          {jsonLdTypes.status === "failed" && (
            <ErrorResult
              error={
                jsonLdTypes.error instanceof z.ZodError
                  ? [
                      "An error occurred while trying to parse the JSON LD Type referenced in this schema:",
                      ...processZodError(jsonLdTypes.error).map((e) => `"${e}"`),
                      "Please provide a schema with a valid JSON LD Type.",
                    ].join("\n")
                  : `An error occurred while downloading the JSON LD Type referenced in this schema:\n"${jsonLdTypes.error}"\nPlease try again.`
              }
            />
          )}

          <Divider />

          <Row justify="end">
            <Button disabled={!jsonLdTypeInput} htmlType="submit" type="primary">
              Preview import
            </Button>
          </Row>
        </Form>
      </Space>
    </Card>
  );
}
