import { Card, Space, message } from "antd";
import { isAxiosError } from "axios";
import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { z } from "zod";

import { createLink } from "src/adapters/api/credentials";
import { getSchemaFromUrl } from "src/adapters/jsonSchemas";
import { credentialFormParser, serializeCredentialForm } from "src/adapters/parsers/forms";
import {
  IssuanceMethodForm,
  IssuanceMethodFormData,
} from "src/components/credentials/IssuanceMethodForm";
import {
  IssueCredentialForm,
  IssueCredentialFormData,
} from "src/components/credentials/IssueCredentialForm";
import { SelectSchema } from "src/components/credentials/SelectSchema";
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

type Step = "issuanceMethod" | "issueCredential" | "summary";

interface Steps {
  issuanceMethod: IssuanceMethodFormData;
  issueCredential: IssueCredentialFormData;
}

const defaultSteps: Steps = {
  issuanceMethod: {
    type: "credentialLink",
  },
  issueCredential: {},
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

  const [step, setStep] = useState<Step>("issuanceMethod");
  const [steps, setSteps] = useState<Steps>(defaultSteps);
  const [schema, setSchema] = useState<Schema>();
  const [jsonSchema, setJsonSchema] = useState<AsyncTask<JsonSchema, string | z.ZodError>>({
    status: "pending",
  });
  const [linkID, setLinkID] = useState<AsyncTask<string, undefined>>({
    status: "pending",
  });

  const { schemaID } = useParams();

  const processError = (error: unknown) =>
    error instanceof z.ZodError ? error : error instanceof Error ? error.message : "Unknown error";

  const onCreateLink = (steps: Steps) => {
    if (schemaID) {
      const parsedForm = credentialFormParser.safeParse(steps);

      if (parsedForm.success) {
        setLinkID({ status: "loading" });
        const serializedCredentialForm = serializeCredentialForm({
          issueCredential: parsedForm.data,
          schemaID,
        });

        if (serializedCredentialForm.success) {
          void createLink({
            env,
            payload: serializedCredentialForm.data,
          }).then((response) => {
            if (response.isSuccessful) {
              setLinkID({ data: response.data.id, status: "successful" });
              setStep("summary");

              void message.success("Credential link created");
            } else {
              setLinkID({ error: undefined, status: "failed" });

              void message.error(response.error.message);
            }
          });
        } else {
          processZodError(serializedCredentialForm.error).forEach((msg) => void message.error(msg));
        }
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

  useEffect(() => {
    if (schemaID) {
      setSteps((currentSteps) => ({
        ...currentSteps,
        issueCredential: defaultSteps.issueCredential,
      }));
    }
  }, [schemaID]);

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
              <IssuanceMethodForm
                initialValues={steps.issuanceMethod}
                onSubmit={(values) => {
                  setSteps({ ...steps, issuanceMethod: values });
                  setStep("issueCredential");
                }}
              />
            );
          }

          case "issueCredential": {
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
                              initialValues={steps.issueCredential}
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
                                const newSteps: Steps =
                                  steps.issuanceMethod.type === "credentialLink" &&
                                  updatedValues.expirationDate?.isBefore(
                                    steps.issuanceMethod.linkExpirationDate
                                  )
                                    ? {
                                        ...steps,
                                        issuanceMethod: {
                                          ...steps.issuanceMethod,
                                          linkExpirationDate: undefined,
                                          linkExpirationTime: undefined,
                                        },
                                        issueCredential: updatedValues,
                                      }
                                    : { ...steps, issueCredential: updatedValues };

                                setSteps(newSteps);
                                onCreateLink(newSteps);
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
            return isAsyncTaskDataAvailable(linkID) && <Summary linkID={linkID.data} />;
          }
        }
      })()}
    </SiderLayoutContent>
  );
}
