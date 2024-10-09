import { Button, Card, Space, Typography } from "antd";
import { useCallback, useEffect, useState } from "react";
import { generatePath, useNavigate, useParams } from "react-router-dom";

import { getApiSchema } from "src/adapters/api/schemas";
import { getJsonSchemaFromUrl, getSchemaJsonLdTypes } from "src/adapters/jsonSchemas";
import CreditCardIcon from "src/assets/icons/credit-card-plus.svg?react";
import { DownloadSchema } from "src/components/schemas/DownloadSchema";
import { SchemaViewer } from "src/components/schemas/SchemaViewer";
import { Detail } from "src/components/shared/Detail";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { ApiSchema, AppError, Json, JsonLdType, JsonSchema } from "src/domain";
import { ROUTES } from "src/routes";
import { AsyncTask, hasAsyncTaskFailed, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { SCHEMA_SEARCH_PARAM } from "src/utils/constants";
import {
  buildAppError,
  jsonLdContextErrorToString,
  jsonSchemaErrorToString,
} from "src/utils/error";
import { formatDate } from "src/utils/forms";

export function SchemaDetails() {
  const { identifier } = useIdentityContext();
  const navigate = useNavigate();
  const { schemaID } = useParams();

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

  const fetchApiSchema = useCallback(
    async (signal: AbortSignal) => {
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
        } else {
          if (!isAbortedError(response.error)) {
            setSchema({ error: response.error, status: "failed" });
          }
        }
      }
    },
    [env, fetchJsonSchemaFromUrl, schemaID, identifier]
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
    isAsyncTaskStarting(contextTuple);

  if (!schemaID) {
    return <ErrorResult error="No schema ID provided." />;
  }

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
          const { bigInt, createdAt, hash, url, version } = schema.data;
          const [jsonSchema, jsonSchemaObject] = jsonSchemaTuple.data;
          const [jsonLdType, jsonLdContextObject] = contextTuple.data;

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

                  <Detail copyable href={url} label="URL" text={url} />

                  <Detail label="Import date" text={formatDate(createdAt)} />

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
      })()}
    </SiderLayoutContent>
  );
}
