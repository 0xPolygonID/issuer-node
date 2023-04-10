import { Card, Space, message } from "antd";
import { isAxiosError } from "axios";
import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";

import { z } from "zod";
import { OldCredential, credentialIssue } from "src/adapters/api/credentials";
import { getSchemaFromUrl } from "src/adapters/jsonSchemas";
import { getIssueCredentialFormDataParser } from "src/adapters/parsers/forms";
import { serializeCredentialForm } from "src/adapters/parsers/serializers";
import {
  AttributeValues,
  IssueCredentialForm,
} from "src/components/credentials/IssueCredentialForm";
import { SelectSchema } from "src/components/credentials/SelectSchema";
import { IssuanceMethod, SetIssuanceMethod } from "src/components/credentials/SetIssuanceMethod";
import { Summary } from "src/components/credentials/Summary";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/env";
import { JsonSchema, Schema } from "src/domain";
import { AsyncTask, isAsyncTaskDataAvailable } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { ISSUE_CREDENTIAL } from "src/utils/constants";
import { processZodError } from "src/utils/error";

type FormStep = "issuanceMethod" | "attributes" | "summary";

interface FormData {
  attributes: AttributeValues;
  issuanceMethod: IssuanceMethod;
}

const defaultFormData: FormData = {
  attributes: {},
  issuanceMethod: {
    type: "credentialLink",
  },
};

const jsonSchemaErrorToString = (error: string | z.ZodError) =>
  error instanceof z.ZodError
    ? [
        "An error occurred while parsing the json schema:",
        ...processZodError(error).map((e) => `"${e}"`),
      ].join("\n")
    : `An error occurred while downloading the json schema from the URL:\n"${error}"\nPlease try again.`;

export function IssueCredential() {
  const env = useEnvContext();

  const [step, setStep] = useState<FormStep>("issuanceMethod");
  const [formData, setFormData] = useState<FormData>(defaultFormData);
  const [schema, setSchema] = useState<Schema>();
  const [jsonSchema, setJsonSchema] = useState<AsyncTask<JsonSchema, string | z.ZodError>>({
    status: "pending",
  });
  const [credential, setCredential] = useState<AsyncTask<OldCredential, undefined>>({
    status: "pending",
  });

  const { schemaID } = useParams();

  const processError = (error: unknown) =>
    error instanceof z.ZodError ? error : error instanceof Error ? error.message : "Unknown error";

  const issueCredential = (formData: FormData) => {
    if (schemaID) {
      const parsedForm = getIssueCredentialFormDataParser([]).safeParse(formData);

      if (parsedForm.success) {
        setCredential({ status: "loading" });

        void credentialIssue({
          env,
          payload: serializeCredentialForm(parsedForm.data),
          schemaID,
        }).then((response) => {
          if (response.isSuccessful) {
            setCredential({ data: response.data, status: "successful" });
            setStep("summary");

            void message.success("Credential link created");
          } else {
            setCredential({ error: undefined, status: "failed" });

            void message.error(response.error.message);
          }
        });
      } else {
        processZodError(parsedForm.error).forEach((msg) => void message.error(msg));
      }
    }
  };

  const fetchJsonSchema = useCallback(
    (signal: AbortSignal) => {
      if (schema) {
        setJsonSchema({ status: "loading" });
        getSchemaFromUrl({
          signal,
          url: schema.url,
        })
          .then(([jsonSchema]) => {
            setJsonSchema({
              data: jsonSchema,
              status: "successful",
            });
          })
          .catch((error) => {
            if (!isAxiosError(error) || !isAbortedError(error)) {
              setJsonSchema({ error: processError(error), status: "failed" });
            }
          });
      }
    },
    [schema]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchJsonSchema);

    return aborter;
  }, [fetchJsonSchema]);

  return (
    <SiderLayoutContent
      description="A credential is issued with assigned attribute values and can be issued directly to identifier or as a credential link containing a QR code."
      showBackButton
      showDivider
      title={ISSUE_CREDENTIAL}
    >
      {(() => {
        switch (step) {
          case "issuanceMethod": {
            return (
              <SetIssuanceMethod
                initialValues={formData.issuanceMethod}
                onSubmit={(values) => {
                  setFormData({ ...formData, issuanceMethod: values });
                  setStep("attributes");
                }}
              />
            );
          }

          case "attributes": {
            return (
              <Card className="issue-credential-card" title="Credential details">
                <Space direction="vertical">
                  <SelectSchema onSelect={setSchema} schemaID={schemaID} />

                  {schema &&
                    (() => {
                      switch (jsonSchema.status) {
                        case "pending":
                        case "loading":
                        case "reloading": {
                          return <LoadingResult />;
                        }

                        case "failed": {
                          return <ErrorResult error={jsonSchemaErrorToString(jsonSchema.error)} />;
                        }

                        case "successful": {
                          return (
                            <IssueCredentialForm
                              initialValues={formData.attributes}
                              jsonSchema={jsonSchema.data}
                              onBack={() => {
                                setJsonSchema({ status: "pending" });
                                setStep("issuanceMethod");
                              }}
                              onSubmit={(values) => {
                                const updatedValues = values.expirationDate
                                  ? {
                                      ...values,
                                      expirationDate: values.expirationDate.endOf("day"),
                                    }
                                  : values;
                                const newFormData: FormData =
                                  formData.issuanceMethod.type === "credentialLink" &&
                                  updatedValues.expirationDate?.isBefore(
                                    formData.issuanceMethod.linkExpirationDate
                                  )
                                    ? {
                                        ...formData,
                                        attributes: updatedValues,
                                        issuanceMethod: {
                                          ...formData.issuanceMethod,
                                          linkExpirationDate: undefined,
                                          linkExpirationTime: undefined,
                                        },
                                      }
                                    : { ...formData, attributes: updatedValues };

                                setFormData(newFormData);
                                setStep("summary");
                              }}
                              schema={schema}
                            />
                          );
                        }
                      }
                    })()}
                </Space>
              </Card>
            );
          }

          case "summary": {
            return (
              schema &&
              isAsyncTaskDataAvailable(credential) && (
                <Summary credential={credential.data} schema={schema} />
              )
            );
          }
        }
      })()}
    </SiderLayoutContent>
  );
}
