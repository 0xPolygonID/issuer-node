import { Button, Card, Row, Space, Typography, message } from "antd";
import { useCallback, useEffect, useState } from "react";
import { generatePath, useNavigate, useParams } from "react-router-dom";
import { z } from "zod";

import { Schema as ApiSchema, getSchema } from "src/adapters/api/schemas";
import { ReactComponent as CreditCardIcon } from "src/assets/icons/credit-card-plus.svg";
import { Detail } from "src/components/schemas/Detail";
import { SchemaViewer } from "src/components/schemas/SchemaViewer";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
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
import {
  AsyncTask,
  hasAsyncTaskFailed,
  isAsyncTaskDataAvailable,
  isAsyncTaskStarting,
} from "src/utils/types";

export function SchemaDetails() {
  const navigate = useNavigate();
  const { schemaID } = useParams();

  const env = useEnvContext();

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
        if (
          isAsyncTaskDataAvailable(apiSchema) &&
          isAsyncTaskDataAvailable(schema) &&
          isAsyncTaskDataAvailable(jsonLdType) &&
          rawJsonLdContext &&
          rawJsonSchema
        ) {
          const { bigInt, createdAt, hash, url } = apiSchema.data;

          return (
            <SchemaViewer
              actions={
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
              }
              contents={
                <Space direction="vertical">
                  <Typography.Text type="secondary">SCHEMA DETAILS</Typography.Text>

                  <Detail copyable data={bigInt} label="BigInt" />

                  <Detail copyable data={hash} label="Hash" />

                  <Detail copyable data={url} label="URL" />

                  <Detail data={formatDate(createdAt, true)} label="Import date" />

                  <Row justify="space-between">
                    <Typography.Text type="secondary">Download</Typography.Text>

                    <Button
                      onClick={() => {
                        downloadJsonFromUrl({
                          fileName: schema.data.name,
                          url: url,
                        })
                          .then(() => {
                            void message.success("Schema downloaded successfully.");
                          })
                          .catch(() => {
                            void message.error(
                              "An error occurred while downloading the schema. Please try again."
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
              jsonLdType={jsonLdType.data}
              rawJsonLdContext={rawJsonLdContext}
              rawJsonSchema={rawJsonSchema}
              schema={schema.data}
            />
          );
        }

        return (
          <Card style={{ margin: "auto", maxWidth: CARD_WIDTH }}>
            {(() => {
              if (hasAsyncTaskFailed(apiSchema)) {
                return (
                  <ErrorResult
                    error={[
                      "An error occurred while downloading or parsing the schema from the API:",
                      apiSchema.error.message,
                    ].join("\n")}
                  />
                );
              } else if (hasAsyncTaskFailed(schema)) {
                return <ErrorResult error={schemaErrorToString(schema.error)} />;
              } else if (hasAsyncTaskFailed(jsonLdType)) {
                return <ErrorResult error={jsonLdTypeErrorToString(jsonLdType.error)} />;
              } else if (loading) {
                return <LoadingResult />;
              }

              return <ErrorResult error="Unknown error" />;
            })()}
          </Card>
        );
      })()}
    </SiderLayoutContent>
  );
}
