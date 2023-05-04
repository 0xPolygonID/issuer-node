import { Button, Card, Divider, Form, Input, Radio, Row, Space } from "antd";
import { useState } from "react";
import { z } from "zod";

import { getJsonSchemaFromUrl, getSchemaJsonLdTypes } from "src/adapters/jsonSchemas";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { AppError, Json, JsonLdType, JsonSchema } from "src/domain";
import { AsyncTask, isAsyncTaskDataAvailable } from "src/utils/async";
import {
  buildAppError,
  jsonLdContextErrorToString,
  jsonSchemaErrorToString,
} from "src/utils/error";

export type FormData = {
  jsonLdType: JsonLdType;
  jsonLdTypes: AsyncTask<JsonLdType[], AppError>;
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
  const [jsonLdTypeInput, setJsonLdTypeInput] = useState<JsonLdType | undefined>(
    initialFormData?.jsonLdType
  );
  const [rawJsonLdContext, setRawJsonLdContext] = useState<Json | undefined>(
    initialFormData?.rawJsonLdContext
  );
  const [rawJsonSchema, setRawJsonSchema] = useState<Json | undefined>(
    initialFormData?.rawJsonSchema
  );
  const [jsonSchema, setJsonSchema] = useState<AsyncTask<JsonSchema, AppError>>(
    initialFormData
      ? { data: initialFormData.jsonSchema, status: "successful" }
      : {
          status: "pending",
        }
  );
  const [jsonLdTypes, setJsonLdTypes] = useState<AsyncTask<JsonLdType[], AppError>>(
    initialFormData
      ? initialFormData.jsonLdTypes
      : {
          status: "pending",
        }
  );

  const fetchJsonSchemaFromUrl = (url: string): void => {
    setJsonSchema({ status: "loading" });

    void getJsonSchemaFromUrl({
      url,
    }).then((jsonSchemaResponse) => {
      if (jsonSchemaResponse.success) {
        const [jsonSchema, rawJsonSchema] = jsonSchemaResponse.data;
        setSchemaUrl(url);
        setJsonSchema({ data: jsonSchema, status: "successful" });
        setRawJsonSchema(rawJsonSchema);
        setJsonLdTypes({ status: "loading" });

        void getSchemaJsonLdTypes({
          jsonSchema,
        }).then((jsonLdTypesResponse) => {
          if (jsonLdTypesResponse.success) {
            const [jsonLdTypes, rawJsonLdContext] = jsonLdTypesResponse.data;
            setJsonLdTypes({ data: jsonLdTypes, status: "successful" });
            setRawJsonLdContext(rawJsonLdContext);

            if (jsonLdTypes.length === 1) {
              setJsonLdTypeInput(jsonLdTypes[0]);
            }
          } else {
            setJsonLdTypes({ error: jsonLdTypesResponse.error, status: "failed" });
          }
        });
      } else {
        setJsonSchema({ error: jsonSchemaResponse.error, status: "failed" });
      }
    });
  };

  const loadSchema = () => {
    const parsedUrl = z.string().safeParse(schemaUrlInput);

    if (parsedUrl.success) {
      setJsonSchema({ status: "pending" });
      setJsonLdTypes({ status: "pending" });
      setJsonLdTypeInput(undefined);
      fetchJsonSchemaFromUrl(parsedUrl.data);
    } else {
      setJsonSchema({
        error: buildAppError(`"${schemaUrlInput}" is not a valid URL`),
        status: "failed",
      });
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
            <ErrorResult error={jsonSchemaErrorToString(jsonSchema.error)} />
          )}

          {jsonLdTypes.status === "failed" && (
            <ErrorResult error={jsonLdContextErrorToString(jsonLdTypes.error)} />
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
