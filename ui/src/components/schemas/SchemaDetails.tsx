import { Button, Card, Dropdown, Row, Space, Typography, message } from "antd";
import { useCallback, useEffect, useState } from "react";
import { generatePath, useNavigate, useParams } from "react-router-dom";
import SyntaxHighlighter from "react-syntax-highlighter";
import { a11yLight } from "react-syntax-highlighter/dist/esm/styles/hljs";

import { z } from "zod";
import { ErrorResult } from "../shared/ErrorResult";
import { LoadingResult } from "../shared/LoadingResult";
import { SiderLayoutContent } from "../shared/SiderLayoutContent";
import { Schema as ApiSchema, getSchema } from "src/adapters/api/schemas";
import { ReactComponent as ChevronDownIcon } from "src/assets/icons/chevron-down.svg";
import { ReactComponent as CreditCardIcon } from "src/assets/icons/credit-card-plus.svg";
import { Detail } from "src/components/schemas/Detail";
import { SchemaTree } from "src/components/schemas/SchemaTree";
import { useEnvContext } from "src/contexts/env";
import { Json, JsonLdType, Schema } from "src/domain";
import { ROUTES } from "src/routes";
import {
  APIError,
  downloadJsonFromUrl,
  getJsonLdTypesFromUrl,
  getSchemaFromUrl,
  processZodError,
} from "src/utils/adapters";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { CARD_WIDTH } from "src/utils/constants";
import { formatDate } from "src/utils/forms";
import { AsyncTask, hasAsyncTaskFailed, isAsyncTaskStarting } from "src/utils/types";

type JsonView = "formatted" | "jsonLdContext" | "jsonSchema";

const JSON_VIEW_LABELS: Record<JsonView, string> = {
  formatted: "Formatted",
  jsonLdContext: "JSON LD Context",
  jsonSchema: "JSON Schema",
};

export function SchemaDetails() {
  const env = useEnvContext();

  const navigate = useNavigate();

  const { schemaID } = useParams();

  const [jsonView, setJsonView] = useState<JsonView>("formatted");
  const [schema, setSchema] = useState<AsyncTask<Schema, string | z.ZodError>>({
    status: "pending",
  });
  const [apiSchema, setApiSchema] = useState<AsyncTask<ApiSchema, APIError>>({
    status: "pending",
  });
  const [rawJsonSchema, setRawJsonSchema] = useState<Json | undefined>();
  const [jsonLdType, setJsonLdType] = useState<AsyncTask<JsonLdType, string | z.ZodError>>({
    status: "pending",
  });
  const [rawJsonLdContext, setRawJsonLdContext] = useState<Json | undefined>();

  const extractError = (error: unknown) =>
    error instanceof z.ZodError ? error : error instanceof Error ? error.message : "Unknown error";

  const fetchSchemaFromUrl = useCallback((apiSchema: ApiSchema): void => {
    setSchema({ status: "loading" });
    getSchemaFromUrl({
      url: apiSchema.url,
    })
      .then(([schema, rawSchema]) => {
        setSchema({ data: schema, status: "successful" });
        setRawJsonSchema(rawSchema);
        setJsonLdType({ status: "loading" });
        getJsonLdTypesFromUrl({
          schema: schema,
          url: schema.$metadata.uris.jsonLdContext,
        })
          .then(([jsonLdTypes, rawJsonLdContext]) => {
            setRawJsonLdContext(rawJsonLdContext);
            const jsonLdType = jsonLdTypes.find((type) => type.name === apiSchema.type);
            if (jsonLdType) {
              setJsonLdType({ data: jsonLdType, status: "successful" });
            } else {
              setJsonLdType({
                error:
                  "Couldn't find the type specified by the schemas API in the context of the schema obtained from the URL",
                status: "failed",
              });
            }
          })
          .catch((error) => {
            setJsonLdType({
              error: extractError(error),
              status: "failed",
            });
          });
      })
      .catch((error) => {
        setSchema({
          error: extractError(error),
          status: "failed",
        });
      });
  }, []);

  const fetchApiSchema = useCallback(
    async (signal: AbortSignal) => {
      if (schemaID) {
        setApiSchema({ status: "loading" });

        const response = await getSchema({
          env,
          schemaID,
          signal,
        });

        if (response.isSuccessful) {
          setApiSchema({ data: response.data, status: "successful" });
          fetchSchemaFromUrl(response.data);
        } else {
          if (!isAbortedError(response.error)) {
            setApiSchema({ error: response.error, status: "failed" });
          }
        }
      }
    },
    [env, fetchSchemaFromUrl, schemaID]
  );

  useEffect(() => {
    if (schemaID) {
      const { aborter } = makeRequestAbortable(fetchApiSchema);
      return aborter;
    }
    return;
  }, [fetchApiSchema, schemaID]);

  const jsonLdTypeErrorToString = (error: string | z.ZodError) =>
    error instanceof z.ZodError
      ? [
          "An error occurred while parsing the context referenced in the schema:",
          ...processZodError(error).map((e) => `"${e}"`),
          "Please provide a schema with a valid context.",
        ].join("\n")
      : `An error occurred while downloading the context referenced in the schema:\n"${error}"\nPlease try again.`;

  const schemaErrorToString = (error: string | z.ZodError) =>
    error instanceof z.ZodError
      ? [
          "An error occurred while parsing the schema from the URL:",
          ...processZodError(error).map((e) => `"${e}"`),
          "Please provide a valid schema.",
        ].join("\n")
      : `An error occurred while downloading the schema from the URL:\n"${error}"\nPlease try again.`;

  const loading =
    isAsyncTaskStarting(apiSchema) ||
    isAsyncTaskStarting(schema) ||
    isAsyncTaskStarting(jsonLdType);

  return (
    <SiderLayoutContent
      description="Schema details include a hash, schema URL and attributes. Schema can be viewed in a formatted way as well as, LD Context and schema."
      showBackButton
      showDivider
      title="Schema details"
    >
      {(() => {
        if (hasAsyncTaskFailed(apiSchema)) {
          return (
            <Card style={{ margin: "auto", maxWidth: CARD_WIDTH }}>
              <ErrorResult
                error={[
                  "An error occurred while downloading or parsing the schema from the API:",
                  apiSchema.error.message,
                ].join("\n")}
              />
            </Card>
          );
        } else if (hasAsyncTaskFailed(schema)) {
          return (
            <Card style={{ margin: "auto", maxWidth: CARD_WIDTH }}>
              <ErrorResult error={schemaErrorToString(schema.error)} />
            </Card>
          );
        } else if (hasAsyncTaskFailed(jsonLdType)) {
          return (
            <Card style={{ margin: "auto", maxWidth: CARD_WIDTH }}>
              <ErrorResult error={jsonLdTypeErrorToString(jsonLdType.error)} />
            </Card>
          );
        } else if (loading) {
          return (
            <Card style={{ margin: "auto", maxWidth: CARD_WIDTH }}>
              <LoadingResult />
            </Card>
          );
        } else {
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
              title={jsonLdType.data.name}
            >
              <Space direction="vertical" size="large">
                <Card className="background-grey">
                  <Space direction="vertical">
                    <Typography.Text type="secondary">SCHEMA DETAILS</Typography.Text>

                    <Detail copyable data={apiSchema.data.bigInt} label="BigInt" />

                    <Detail copyable data={apiSchema.data.hash} label="Hash" />

                    <Detail copyable data={apiSchema.data.url} label="URL" />

                    <Detail data={formatDate(apiSchema.data.createdAt, true)} label="Import date" />

                    <Row justify="space-between">
                      <Typography.Text type="secondary">Download</Typography.Text>

                      <Button
                        onClick={() => {
                          downloadJsonFromUrl({
                            fileName: schema.data.name,
                            url: apiSchema.data.url,
                          })
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
                </Card>

                {(() => {
                  switch (jsonView) {
                    case "formatted": {
                      return (
                        <Card className="background-grey">
                          <Space direction="vertical">
                            <Typography.Text type="secondary">ATTRIBUTES</Typography.Text>

                            <SchemaTree className="background-grey" schema={schema.data} />
                          </Space>
                        </Card>
                      );
                    }
                    case "jsonLdContext": {
                      return (
                        <SyntaxHighlighter
                          className="background-grey"
                          customStyle={{
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
                          className="background-grey"
                          customStyle={{
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
                    <Button
                      icon={<CreditCardIcon />}
                      onClick={() => {
                        navigate(generatePath(ROUTES.issueCredential.path, { schemaID }));
                      }}
                      type="primary"
                    >
                      Issue credential
                    </Button>
                  </Space>
                </Row>
              </Space>
            </Card>
          );
        }
      })()}
    </SiderLayoutContent>
  );
}
