import { Card, Space, message } from "antd";
import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";

import { APIError } from "src/adapters/api";
import { OldCredential, credentialIssue } from "src/adapters/api/credentials";
import { getSchema } from "src/adapters/api/schemas";
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
import { Schema } from "src/domain";
import { AsyncTask, isAsyncTaskDataAvailable } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { ISSUE_CREDENTIAL } from "src/utils/constants";
import { processZodError } from "src/utils/error";

type IssuanceStep = "attributes" | "issuanceMethod" | "summary";

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

export function IssueCredential() {
  const env = useEnvContext();
  const [credential, setCredential] = useState<AsyncTask<OldCredential, undefined>>({
    status: "pending",
  });
  const [formData, setFormData] = useState<FormData>(defaultFormData);
  const [schema, setSchema] = useState<AsyncTask<Schema, APIError>>({
    status: "pending",
  });

  const [step, setStep] = useState<IssuanceStep>("attributes");

  const { schemaID } = useParams();

  const fetchSchema = useCallback(
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
        } else {
          if (!isAbortedError(response.error)) {
            setSchema({ error: response.error, status: "failed" });
          }
        }
      }
    },
    [env, schemaID]
  );

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

  useEffect(() => {
    setStep("attributes");
    setFormData(defaultFormData);

    if (schemaID) {
      const { aborter } = makeRequestAbortable(fetchSchema);

      return aborter;
    } else {
      setSchema({ status: "pending" });
    }
    return;
  }, [fetchSchema, schemaID]);

  return (
    <SiderLayoutContent
      description="A credential is issued with assigned attribute values and can be issued directly to identifier or as a credential link containing a QR code."
      showBackButton
      showDivider
      title={ISSUE_CREDENTIAL}
    >
      {(() => {
        switch (schema.status) {
          case "failed": {
            return <ErrorResult error={schema.error.message} />;
          }
          case "loading": {
            return <LoadingResult />;
          }
          case "reloading":
          case "pending":
          case "successful": {
            switch (step) {
              case "attributes": {
                return (
                  <Card className="issue-credential-card" title="Credential details">
                    <Space direction="vertical">
                      <SelectSchema schemaID={schemaID} />

                      <IssueCredentialForm
                        initialValues={formData.attributes}
                        onSubmit={(values) => {
                          const updatedValues = values.expirationDate
                            ? { ...values, expirationDate: values.expirationDate.endOf("day") }
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
                          setStep("issuanceMethod");
                        }}
                        schema={isAsyncTaskDataAvailable(schema) ? schema.data : undefined}
                      />
                    </Space>
                  </Card>
                );
              }

              case "issuanceMethod": {
                return (
                  isAsyncTaskDataAvailable(schema) && (
                    <SetIssuanceMethod
                      credentialExpirationDate={formData.attributes.expirationDate}
                      initialValues={formData.issuanceMethod}
                      isCredentialLoading={credential.status === "loading"}
                      onStepBack={(values) => {
                        const newFormData = { ...formData, issuanceMethod: values };

                        setFormData(newFormData);
                        setStep("attributes");
                      }}
                      onSubmit={(values) => {
                        const newFormData = { ...formData, issuanceMethod: values };

                        setFormData(newFormData);
                        issueCredential(newFormData);
                      }}
                    />
                  )
                );
              }

              case "summary": {
                return (
                  isAsyncTaskDataAvailable(schema) &&
                  isAsyncTaskDataAvailable(credential) && (
                    <Summary credential={credential.data} schema={schema.data} />
                  )
                );
              }
            }
          }
        }
      })()}
    </SiderLayoutContent>
  );
}
