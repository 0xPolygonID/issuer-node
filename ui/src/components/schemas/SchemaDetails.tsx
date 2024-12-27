import { App, Button, Card, Space, Typography } from "antd";
import { useCallback, useEffect, useState } from "react";
import { generatePath, useNavigate, useParams } from "react-router-dom";
import { getDisplayMethods } from "src/adapters/api/display-method";

import { UpdateSchema, getApiSchema, processUrl, updateSchema } from "src/adapters/api/schemas";
import { getJsonSchemaFromUrl, getSchemaJsonLdTypes } from "src/adapters/jsonSchemas";
import {
  buildAppError,
  jsonLdContextErrorToString,
  jsonSchemaErrorToString,
  notifyErrors,
} from "src/adapters/parsers";
import CreditCardIcon from "src/assets/icons/credit-card-plus.svg?react";
import { DownloadSchema } from "src/components/schemas/DownloadSchema";
import { SchemaViewer } from "src/components/schemas/SchemaViewer";
import { Detail } from "src/components/shared/Detail";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { ApiSchema, AppError, DisplayMethod, Json, JsonLdType, JsonSchema } from "src/domain";
import { ROUTES } from "src/routes";
import {
  AsyncTask,
  hasAsyncTaskFailed,
  isAsyncTaskDataAvailable,
  isAsyncTaskStarting,
} from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { SCHEMA_SEARCH_PARAM } from "src/utils/constants";
import { formatDate } from "src/utils/forms";

export function SchemaDetails() {
  const { identifier } = useIdentityContext();
  const navigate = useNavigate();
  const { schemaID } = useParams();
  const { message } = App.useApp();
  const env = useEnvContext();

  const [jsonSchemaTuple, setJsonSchemaTuple] = useState<AsyncTask<[JsonSchema, Json], AppError>>({
    status: "pending",
  });
  const [schema, setSchema] = useState<AsyncTask<ApiSchema, AppError>>({
    status: "pending",
  });
  const [contextTuple, setContextTuple] = useState<AsyncTask<[JsonLdType, Json], AppError>>({
    status: "pending",
  });

  const [displayMethods, setDisplayMethods] = useState<AsyncTask<DisplayMethod[], AppError>>({
    status: "pending",
  });

  const fetchJsonSchemaFromUrl = useCallback(
    (schema: ApiSchema): void => {
      setJsonSchemaTuple({ status: "loading" });

      void getJsonSchemaFromUrl({
        env,
        url: schema.url,
      }).then((jsonSchemaResponse) => {
        if (jsonSchemaResponse.success) {
          const [jsonSchema, jsonSchemaObject] = jsonSchemaResponse.data;
          setJsonSchemaTuple({ data: [jsonSchema, jsonSchemaObject], status: "successful" });
          setContextTuple({ status: "loading" });
          void getSchemaJsonLdTypes({
            env,
            jsonSchema,
          }).then((jsonLdTypesResponse) => {
            if (jsonLdTypesResponse.success) {
              const [jsonLdTypes, jsonLdContextObject] = jsonLdTypesResponse.data;
              const jsonLdType = jsonLdTypes.find((type) => type.name === schema.type);

              if (jsonLdType) {
                setContextTuple({ data: [jsonLdType, jsonLdContextObject], status: "successful" });
              } else {
                setContextTuple({
                  error: buildAppError(
                    "Couldn't find the type specified by the schemas API in the context of the schema obtained from the URL"
                  ),
                  status: "failed",
                });
              }
            } else {
              setContextTuple({
                error: jsonLdTypesResponse.error,
                status: "failed",
              });
            }
          });
        } else {
          setJsonSchemaTuple({
            error: jsonSchemaResponse.error,
            status: "failed",
          });
        }
      });
    },
    [env]
  );

  const fetchDisplayMethods = useCallback(async () => {
    setDisplayMethods((previousDisplayMethods) =>
      isAsyncTaskDataAvailable(previousDisplayMethods)
        ? { data: previousDisplayMethods.data, status: "reloading" }
        : { status: "loading" }
    );

    const response = await getDisplayMethods({
      env,
      identifier,
      params: {},
    });
    if (response.success) {
      setDisplayMethods({
        data: response.data.items.successful,
        status: "successful",
      });

      void notifyErrors(response.data.items.failed);
    } else {
      if (!isAbortedError(response.error)) {
        setDisplayMethods({ error: response.error, status: "failed" });
      }
    }
  }, [env, identifier]);

  const fetchApiSchema = useCallback(
    async (signal?: AbortSignal) => {
      if (schemaID) {
        setSchema({ status: "loading" });

        const response = await getApiSchema({
          env,
          identifier,
          schemaID,
          signal,
        });

        if (response.success) {
          setSchema({ data: response.data, status: "successful" });
          fetchJsonSchemaFromUrl(response.data);
          void fetchDisplayMethods();
        } else {
          if (!isAbortedError(response.error)) {
            setSchema({ error: response.error, status: "failed" });
          }
        }
      }
    },
    [env, fetchJsonSchemaFromUrl, schemaID, identifier, fetchDisplayMethods]
  );

  useEffect(() => {
    if (schemaID) {
      const { aborter } = makeRequestAbortable(fetchApiSchema);
      return aborter;
    }
    return;
  }, [fetchApiSchema, schemaID]);

  const loading =
    isAsyncTaskStarting(schema) ||
    isAsyncTaskStarting(jsonSchemaTuple) ||
    isAsyncTaskStarting(contextTuple) ||
    isAsyncTaskStarting(displayMethods);

  if (!schemaID) {
    return <ErrorResult error="No schema ID provided." />;
  }

  const handleEdit = (formValues: UpdateSchema) => {
    void updateSchema({
      env,
      identifier,
      payload: formValues,
      schemaID,
    }).then((response) => {
      if (response.success) {
        void fetchApiSchema().then(() => {
          void message.success("Schema edited successfully");
        });
      } else {
        void message.error(response.error.message);
      }
    });
  };

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
              <ErrorResult error={jsonSchemaErrorToString(jsonSchemaTuple.error)} />
            </Card>
          );
        } else if (hasAsyncTaskFailed(contextTuple)) {
          return (
            <Card className="centered">
              <ErrorResult error={jsonLdContextErrorToString(contextTuple.error)} />
            </Card>
          );
        } else if (loading) {
          return (
            <Card className="centered">
              <LoadingResult />
            </Card>
          );
        } else {
          const { bigInt, createdAt, displayMethodID, hash, url, version } = schema.data;
          const processedSchemaUrl = processUrl(url, env);
          const [jsonSchema, jsonSchemaObject] = jsonSchemaTuple.data;
          const [jsonLdType, jsonLdContextObject] = contextTuple.data;
          const displayMethodsList = isAsyncTaskDataAvailable(displayMethods)
            ? displayMethods.data
            : [];
          const displayMethod = displayMethodsList.find(({ id }) => id === displayMethodID);

          return (
            <SchemaViewer
              actions={
                <Space size="middle">
                  <Button
                    icon={<CreditCardIcon />}
                    onClick={() => {
                      navigate({
                        pathname: generatePath(ROUTES.issueCredential.path),
                        search: `${SCHEMA_SEARCH_PARAM}=${schemaID}`,
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

                  <Detail copyable label="Schema type" text={jsonLdType.name} />

                  {version && <Detail label="Schema version" text={version} />}

                  <Detail copyable label="BigInt" text={bigInt} />

                  <Detail copyable label="Hash" text={hash} />

                  <Detail
                    copyable
                    href={processedSchemaUrl.success ? processedSchemaUrl.data : url}
                    label="URL"
                    text={url}
                  />

                  {displayMethod && (
                    <Detail
                      copyable
                      href={generatePath(ROUTES.displayMethodDetails.path, { displayMethodID })}
                      label="Default display method"
                      text={displayMethod.name}
                    />
                  )}

                  <Detail label="Import date" text={formatDate(createdAt)} />

                  <DownloadSchema env={env} fileName={jsonSchema.name} url={url} />
                </Space>
              }
              displayMethodID={displayMethodID}
              displayMethods={displayMethodsList}
              jsonLdContextObject={jsonLdContextObject}
              jsonLdType={jsonLdType}
              jsonSchema={jsonSchema}
              jsonSchemaObject={jsonSchemaObject}
              onEdit={handleEdit}
            />
          );
        }
      })()}
    </SiderLayoutContent>
  );
}
