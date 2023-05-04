import { Button, Card, Row, Space, Typography, message } from "antd";
import { useCallback, useEffect, useState } from "react";
import { generatePath, useNavigate, useParams } from "react-router-dom";
import { z } from "zod";

import { APIError } from "src/adapters/api";
import { getSchema } from "src/adapters/api/schemas";
import { downloadJsonFromUrl } from "src/adapters/json";
import { getJsonSchemaFromUrl, getSchemaJsonLdTypes } from "src/adapters/jsonSchemas";
import { ReactComponent as CreditCardIcon } from "src/assets/icons/credit-card-plus.svg";
import { SchemaViewer } from "src/components/schemas/SchemaViewer";
import { Detail } from "src/components/shared/Detail";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { Json, JsonLdType, JsonSchema, Schema } from "src/domain";
import { ROUTES } from "src/routes";
import { AsyncTask, hasAsyncTaskFailed, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { SCHEMA_SEARCH_PARAM } from "src/utils/constants";
import { processError, processZodError } from "src/utils/error";
import { formatDate } from "src/utils/forms";

export function SchemaDetails() {
  const navigate = useNavigate();
  const { schemaID } = useParams();

  const env = useEnvContext();

  const [jsonSchemaTuple, setJsonSchemaTuple] = useState<
    AsyncTask<[JsonSchema, Json], string | z.ZodError>
  >({
    status: "pending",
  });
  const [schema, setSchema] = useState<AsyncTask<Schema, APIError>>({
    status: "pending",
  });
  const [contextTuple, setContextTuple] = useState<
    AsyncTask<[JsonLdType, Json], string | z.ZodError>
  >({
    status: "pending",
  });

  const fetchJsonSchemaFromUrl = useCallback((schema: Schema): void => {
    setJsonSchemaTuple({ status: "loading" });

    getJsonSchemaFromUrl({
      url: schema.url,
    })
      .then(([jsonSchema, rawJsonSchema]) => {
        setJsonSchemaTuple({ data: [jsonSchema, rawJsonSchema], status: "successful" });
        setContextTuple({ status: "loading" });
        getSchemaJsonLdTypes({
          jsonSchema,
        })
          .then(([jsonLdTypes, rawJsonLdContext]) => {
            const jsonLdType = jsonLdTypes.find((type) => type.name === schema.type);

            if (jsonLdType) {
              setContextTuple({ data: [jsonLdType, rawJsonLdContext], status: "successful" });
            } else {
              setContextTuple({
                error:
                  "Couldn't find the type specified by the schemas API in the context of the schema obtained from the URL",
                status: "failed",
              });
            }
          })
          .catch((error) => {
            setContextTuple({
              error: processError(error),
              status: "failed",
            });
          });
      })
      .catch((error) => {
        setJsonSchemaTuple({
          error: processError(error),
          status: "failed",
        });
      });
  }, []);

  const fetchApiSchema = useCallback(
    async (signal: AbortSignal) => {
      if (schemaID) {
        setSchema({ status: "loading" });

        const response = await getSchema({
          env,
          schemaID,
          signal,
        });

        if (response.isSuccessful) {
          setSchema({ data: response.data, status: "successful" });
          fetchJsonSchemaFromUrl(response.data);
        } else {
          if (!isAbortedError(response.error)) {
            setSchema({ error: response.error, status: "failed" });
          }
        }
      }
    },
    [env, fetchJsonSchemaFromUrl, schemaID]
  );

  useEffect(() => {
    if (schemaID) {
      const { aborter } = makeRequestAbortable(fetchApiSchema);
      return aborter;
    }
    return;
  }, [fetchApiSchema, schemaID]);

  const contextTupleErrorToString = (error: string | z.ZodError) =>
    error instanceof z.ZodError
      ? [
          "An error occurred while parsing the context referenced in the schema:",
          ...processZodError(error).map((e) => `"${e}"`),
          "Please provide a schema with a valid context.",
        ].join("\n")
      : `An error occurred while downloading the context referenced in the schema:\n"${error}"\nPlease try again.`;

  const jsonSchemaTupleErrorToString = (error: string | z.ZodError) =>
    error instanceof z.ZodError
      ? [
          "An error occurred while parsing the schema from the URL:",
          ...processZodError(error).map((e) => `"${e}"`),
          "Please provide a valid schema.",
        ].join("\n")
      : `An error occurred while downloading the schema from the URL:\n"${error}"\nPlease try again.`;

  const loading =
    isAsyncTaskStarting(schema) ||
    isAsyncTaskStarting(jsonSchemaTuple) ||
    isAsyncTaskStarting(contextTuple);

  return (
    <SiderLayoutContent
      description="Schema details include a hash, schema URL and attributes. The schema can be viewed formatted by its attributes, as the JSON LD Context or as a JSON."
      showBackButton
      showDivider
      title="Schema details"
    >
      {(() => {
        if (hasAsyncTaskFailed(schema)) {
          return (
            <Card className="centered">
              <ErrorResult
                error={[
                  "An error occurred while downloading or parsing the schema from the API:",
                  schema.error.message,
                ].join("\n")}
              />
            </Card>
          );
        } else if (hasAsyncTaskFailed(jsonSchemaTuple)) {
          return (
            <Card className="centered">
              <ErrorResult error={jsonSchemaTupleErrorToString(jsonSchemaTuple.error)} />
            </Card>
          );
        } else if (hasAsyncTaskFailed(contextTuple)) {
          return (
            <Card className="centered">
              <ErrorResult error={contextTupleErrorToString(contextTuple.error)} />
            </Card>
          );
        } else if (loading) {
          return (
            <Card className="centered">
              <LoadingResult />
            </Card>
          );
        } else {
          const { bigInt, createdAt, hash, url } = schema.data;
          const [jsonSchema, rawJsonSchema] = jsonSchemaTuple.data;
          const [jsonLdType, rawJsonLdContext] = contextTuple.data;

          return (
            <SchemaViewer
              actions={
                <Space size="middle">
                  <Button
                    icon={<CreditCardIcon />}
                    onClick={() => {
                      navigate({
                        pathname: generatePath(ROUTES.issueCredential.path),
                        search: schemaID ? `?${SCHEMA_SEARCH_PARAM}=${schemaID}` : "",
                      });
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

                  <Detail copyable label="BigInt" text={bigInt} />

                  <Detail copyable label="Hash" text={hash} />

                  <Detail copyable label="URL" text={url} />

                  <Detail label="Import date" text={formatDate(createdAt)} />

                  <Row justify="space-between">
                    <Typography.Text type="secondary">Download</Typography.Text>

                    <Button
                      onClick={() => {
                        downloadJsonFromUrl({
                          fileName: jsonSchema.name,
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
              jsonLdType={jsonLdType}
              jsonSchema={jsonSchema}
              rawJsonLdContext={rawJsonLdContext}
              rawJsonSchema={rawJsonSchema}
            />
          );
        }
      })()}
    </SiderLayoutContent>
  );
}
