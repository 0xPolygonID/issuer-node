import { Button, Card, Form, Input, Radio, Row, Space } from "antd";
import { useState } from "react";
import { z } from "zod";

import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { Json, JsonLdType, Schema } from "src/domain";
import { getJsonLdTypesFromUrl, getSchemaFromUrl, processZodError } from "src/utils/adapters";
import { CARD_WIDTH } from "src/utils/constants";
import { AsyncTask, isAsyncTaskDataAvailable } from "src/utils/types";

export type FormData = {
  jsonLdType: JsonLdType;
  jsonLdTypes: AsyncTask<JsonLdType[], string | z.ZodError>;
  rawJsonLdContext: Json;
  rawJsonSchema: Json;
  schema: Schema;
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
  const [schema, setSchema] = useState<AsyncTask<Schema, string | z.ZodError>>(
    initialFormData
      ? { data: initialFormData.schema, status: "successful" }
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
    setSchema({ status: "loading" });
    getSchemaFromUrl({
      url,
    })
      .then(([schema, rawSchema]) => {
        setSchemaUrl(url);
        setSchema({ data: schema, status: "successful" });
        setRawJsonSchema(rawSchema);
        setJsonLdTypes({ status: "loading" });
        getJsonLdTypesFromUrl({
          schema: schema,
          url: schema.$metadata.uris.jsonLdContext,
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
        setSchema({
          error: processError(error),
          status: "failed",
        });
      });
  };

  const loadSchema = () => {
    const parsedUrl = z.string().safeParse(schemaUrlInput);
    if (parsedUrl.success) {
      setSchema({ status: "pending" });
      setJsonLdTypes({ status: "pending" });
      setJsonLdTypeInput(undefined);
      fetchSchemaFromUrl(parsedUrl.data);
    } else {
      setSchema({ error: `"${schemaUrlInput}" is not a valid URL`, status: "failed" });
    }
  };

  return (
    <Card style={{ margin: "auto", maxWidth: CARD_WIDTH }} title="Provide JSON schema URL">
      <Form
        layout="vertical"
        onFinish={() => {
          if (
            schemaUrl &&
            isAsyncTaskDataAvailable(schema) &&
            jsonLdTypeInput &&
            rawJsonSchema &&
            rawJsonLdContext
          ) {
            onFinish({
              jsonLdType: jsonLdTypeInput,
              jsonLdTypes: jsonLdTypes,
              rawJsonLdContext,
              rawJsonSchema,
              schema: schema.data,
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

        {(schema.status === "loading" || jsonLdTypes.status === "loading") && <LoadingResult />}

        {schema.status === "failed" && (
          <ErrorResult
            error={
              schema.error instanceof z.ZodError
                ? [
                    "An error occurred while trying to parse this schema:",
                    ...processZodError(schema.error).map((e) => `"${e}"`),
                    "Please provide a valid JSON Schema.",
                  ].join("\n")
                : `An error occurred while downloading this schema:\n"${schema.error}"\nPlease try again.`
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

        <Row justify="end">
          <Button disabled={!jsonLdTypeInput} htmlType="submit" type="primary">
            Preview import
          </Button>
        </Row>
      </Form>
    </Card>
  );
}
