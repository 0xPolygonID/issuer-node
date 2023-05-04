import { Button, Card, Row, Space, message } from "antd";
import { isAxiosError } from "axios";
import { useCallback, useEffect, useState } from "react";
import { generatePath, useNavigate, useSearchParams } from "react-router-dom";
import { z } from "zod";

import { createCredential, createLink } from "src/adapters/api/credentials";
import { getJsonSchemaFromUrl } from "src/adapters/jsonSchemas";
import {
  CredentialDirectIssuance,
  CredentialFormInput,
  CredentialLinkIssuance,
  credentialFormParser,
  serializeCredentialIssuance,
  serializeCredentialLinkIssuance,
} from "src/adapters/parsers/forms";
import { ReactComponent as IconBack } from "src/assets/icons/arrow-narrow-left.svg";
import { ReactComponent as IconRight } from "src/assets/icons/arrow-narrow-right.svg";
import { IssuanceMethodForm } from "src/components/credentials/IssuanceMethodForm";
import { IssueCredentialForm } from "src/components/credentials/IssueCredentialForm";
import { SelectSchema } from "src/components/credentials/SelectSchema";
import { Summary } from "src/components/credentials/Summary";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { useIssuerStateContext } from "src/contexts/IssuerState";
import { JsonSchema, Schema } from "src/domain";
import { ROUTES } from "src/routes";
import { AsyncTask, isAsyncTaskDataAvailable } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import {
  CREDENTIALS_TABS,
  DID_SEARCH_PARAM,
  ISSUE_CREDENTIAL,
  ISSUE_CREDENTIAL_DIRECT,
  ISSUE_CREDENTIAL_LINK,
  SCHEMA_SEARCH_PARAM,
} from "src/utils/constants";
import { processError, processZodError } from "src/utils/error";

type Step = "issuanceMethod" | "issueCredential" | "summary";

const defaultCredentialFormInput: CredentialFormInput = {
  issuanceMethod: {
    type: "directIssue",
  },
  issueCredential: {
    proofTypes: ["SIG"],
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
  const { notifyChange } = useIssuerStateContext();

  const navigate = useNavigate();

  const [searchParams] = useSearchParams();
  const schemaID = searchParams.get(SCHEMA_SEARCH_PARAM) || undefined;
  const did = searchParams.get(DID_SEARCH_PARAM) || undefined;

  const [step, setStep] = useState<Step>(did ? "issueCredential" : "issuanceMethod");
  const [credentialFormInput, setCredentialFormInput] = useState<CredentialFormInput>(
    defaultCredentialFormInput.issuanceMethod.type === "directIssue"
      ? {
          ...defaultCredentialFormInput,
          issuanceMethod: { ...defaultCredentialFormInput.issuanceMethod, did },
        }
      : defaultCredentialFormInput
  );

  const [schema, setSchema] = useState<Schema>();
  const [jsonSchema, setJsonSchema] = useState<AsyncTask<JsonSchema, string | z.ZodError>>({
    status: "pending",
  });
  const [linkID, setLinkID] = useState<AsyncTask<string, null>>({
    status: "pending",
  });
  const [isLoading, setIsLoading] = useState<boolean>(false);

  const onChangeDid = (did: string) => {
    const search = new URLSearchParams(searchParams);

    if (did) {
      search.set(DID_SEARCH_PARAM, did);
    } else {
      search.delete(DID_SEARCH_PARAM);
    }

    navigate(
      {
        pathname: generatePath(ROUTES.issueCredential.path),
        search: search.toString(),
      },
      { replace: true }
    );
  };

  const onChangeSchema = useCallback(
    (schema: Schema) => {
      const search = new URLSearchParams(searchParams);

      search.set(SCHEMA_SEARCH_PARAM, schema.id);

      navigate(
        {
          pathname: generatePath(ROUTES.issueCredential.path),
          search: search.toString(),
        },
        { replace: true }
      );

      setSchema((currentSchema) => (currentSchema?.id === schema.id ? currentSchema : schema));
    },
    [navigate, searchParams]
  );

  const createCredentialLink = async (credentialLinkIssuance: CredentialLinkIssuance) => {
    if (schemaID) {
      setLinkID({ status: "loading" });
      setIsLoading(true);
      const serializedCredentialForm = serializeCredentialLinkIssuance({
        issueCredential: credentialLinkIssuance,
        schemaID,
      });

      if (serializedCredentialForm.success) {
        const response = await createLink({
          env,
          payload: serializedCredentialForm.data,
        });
        if (response.isSuccessful) {
          setLinkID({ data: response.data.id, status: "successful" });
          setStep("summary");

          void message.success("Credential link created");
        } else {
          setLinkID({ error: null, status: "failed" });

          void message.error(response.error.message);
        }
      } else {
        processZodError(serializedCredentialForm.error).forEach((msg) => void message.error(msg));
      }
      setIsLoading(false);
    }
  };

  const issueCredential = async (credentialIssuance: CredentialDirectIssuance) => {
    if (schema) {
      setIsLoading(true);
      const serializedCredentialForm = serializeCredentialIssuance({
        credentialSchema: schema.url,
        issueCredential: credentialIssuance,
        type: schema.type,
      });

      if (serializedCredentialForm.success) {
        const response = await createCredential({
          env,
          payload: serializedCredentialForm.data,
        });
        if (response.isSuccessful) {
          navigate(
            generatePath(ROUTES.credentials.path, {
              tabID: CREDENTIALS_TABS[0].tabID,
            })
          );

          if (credentialIssuance.mtProof) {
            void notifyChange("credential");
          }

          void message.success("Credential issued");
        } else {
          void message.error(response.error.message);
        }
      } else {
        processZodError(serializedCredentialForm.error).forEach((msg) => void message.error(msg));
      }

      setIsLoading(false);
    }
  };

  const fetchJsonSchema = useCallback(
    (signal: AbortSignal) => {
      if (schema && step === "issueCredential") {
        setJsonSchema({ status: "loading" });
        getJsonSchemaFromUrl({
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
    [schema, step]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchJsonSchema);

    return aborter;
  }, [fetchJsonSchema]);

  useEffect(() => {
    if (schemaID) {
      setCredentialFormInput((currentCredentialFormInput) => ({
        ...currentCredentialFormInput,
        issueCredential: defaultCredentialFormInput.issueCredential,
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
                did={did}
                initialValues={credentialFormInput.issuanceMethod}
                onChangeDid={onChangeDid}
                onSubmit={(values) => {
                  if (values.type === "directIssue" && values.did) {
                    onChangeDid(values.did);
                  } else {
                    onChangeDid("");
                  }

                  setCredentialFormInput({ ...credentialFormInput, issuanceMethod: values });
                  setStep("issueCredential");
                }}
              />
            );
          }

          case "issueCredential": {
            const onBack = () => {
              setJsonSchema({ status: "pending" });
              setStep("issuanceMethod");
            };
            return (
              <Card className="issue-credential-card" title="Credential details">
                <Space direction="vertical">
                  <SelectSchema onSelect={onChangeSchema} schemaID={schemaID} />

                  {schema ? (
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
                              initialValues={credentialFormInput.issueCredential}
                              jsonSchema={jsonSchema.data}
                              loading={isLoading}
                              onBack={onBack}
                              onSubmit={(values) => {
                                const newCredentialFormInput: CredentialFormInput = {
                                  ...credentialFormInput,
                                  issueCredential: values,
                                };

                                setCredentialFormInput(newCredentialFormInput);

                                const parsedForm =
                                  credentialFormParser.safeParse(newCredentialFormInput);

                                if (parsedForm.success) {
                                  if (parsedForm.data.type === "credentialLink") {
                                    void createCredentialLink(parsedForm.data);
                                  } else {
                                    void issueCredential(parsedForm.data);
                                  }
                                } else {
                                  processZodError(parsedForm.error).forEach(
                                    (msg) => void message.error(msg)
                                  );
                                }
                              }}
                              schema={schema}
                              type={credentialFormInput.issuanceMethod.type}
                            />
                          );
                        }
                      }
                    })()
                  ) : (
                    <Row justify="end">
                      <Space size="middle">
                        <Button icon={<IconBack />} onClick={onBack} type="default">
                          Previous step
                        </Button>

                        <Button disabled htmlType="submit" type="primary">
                          {credentialFormInput.issuanceMethod.type === "directIssue"
                            ? ISSUE_CREDENTIAL_DIRECT
                            : ISSUE_CREDENTIAL_LINK}
                          <IconRight />
                        </Button>
                      </Space>
                    </Row>
                  )}
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
