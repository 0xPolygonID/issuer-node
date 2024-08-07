import { Card, Space, message } from "antd";
import { useCallback, useEffect, useState } from "react";
import { generatePath, useNavigate, useSearchParams } from "react-router-dom";

import { createCredential, createLink } from "src/adapters/api/credentials";
import {
  CredentialDirectIssuance,
  CredentialFormInput,
  CredentialLinkIssuance,
  IssueCredentialFormData,
  credentialFormParser,
  serializeCredentialIssuance,
  serializeCredentialLinkIssuance,
} from "src/adapters/parsers/view";
import { IssuanceMethodForm } from "src/components/credentials/IssuanceMethodForm";
import { IssueCredentialForm } from "src/components/credentials/IssueCredentialForm";
import { Summary } from "src/components/credentials/Summary";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { useIssuerContext } from "src/contexts/Issuer";
import { useIssuerStateContext } from "src/contexts/IssuerState";
import { ApiSchema, JsonSchema } from "src/domain";
import { ROUTES } from "src/routes";
import { AsyncTask, isAsyncTaskDataAvailable } from "src/utils/async";
import {
  CREDENTIALS_TABS,
  DID_SEARCH_PARAM,
  ISSUE_CREDENTIAL,
  SCHEMA_SEARCH_PARAM,
} from "src/utils/constants";
import { notifyParseError } from "src/utils/error";
import {
  extractCredentialSubjectAttribute,
  extractCredentialSubjectAttributeWithoutId,
} from "src/utils/jsonSchemas";

type Step = "issuanceMethod" | "issueCredential" | "summary";

const defaultCredentialFormInput: CredentialFormInput = {
  issuanceMethod: {
    type: "directIssue",
  },
  issueCredential: {
    proofTypes: ["SIG"],
    refreshService: { enabled: false, url: "" },
  },
};

export function IssueCredential() {
  const env = useEnvContext();
  const { identifier } = useIssuerContext();
  const { notifyChange } = useIssuerStateContext();

  const navigate = useNavigate();
  const [messageAPI, messageContext] = message.useMessage();
  const [searchParams] = useSearchParams();

  const schemaID = searchParams.get(SCHEMA_SEARCH_PARAM) || undefined;
  const did = searchParams.get(DID_SEARCH_PARAM) || undefined;

  const [step, setStep] = useState<Step>(did ? "issueCredential" : "issuanceMethod");
  const [credentialFormInput, setCredentialFormInput] = useState<CredentialFormInput>(
    defaultCredentialFormInput.issuanceMethod.type === "directIssue"
      ? {
          ...defaultCredentialFormInput,
          issuanceMethod: { ...defaultCredentialFormInput.issuanceMethod, did },
          issueCredential: { ...defaultCredentialFormInput.issueCredential, schemaID },
        }
      : {
          ...defaultCredentialFormInput,
          issueCredential: { ...defaultCredentialFormInput.issueCredential, schemaID },
        }
  );

  const [linkID, setLinkID] = useState<AsyncTask<string, null>>({
    status: "pending",
  });
  const [isLoading, setIsLoading] = useState<boolean>(false);

  const onChangeDid = (did?: string) => {
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

  const onSelectApiSchema = useCallback(
    (schema: ApiSchema) => {
      const search = new URLSearchParams(searchParams);

      search.set(SCHEMA_SEARCH_PARAM, schema.id);

      navigate(
        {
          pathname: generatePath(ROUTES.issueCredential.path),
          search: search.toString(),
        },
        { replace: true }
      );
    },
    [navigate, searchParams]
  );

  const createCredentialLink = async ({
    credentialLinkIssuance,
    jsonSchema,
  }: {
    credentialLinkIssuance: CredentialLinkIssuance;
    jsonSchema: JsonSchema;
  }) => {
    const credentialSubjectAttributeWithoutId =
      extractCredentialSubjectAttributeWithoutId(jsonSchema);

    if (schemaID && credentialSubjectAttributeWithoutId) {
      setLinkID({ status: "loading" });
      setIsLoading(true);
      const serializedCredentialForm = serializeCredentialLinkIssuance({
        attribute: credentialSubjectAttributeWithoutId,
        issueCredential: credentialLinkIssuance,
        schemaID,
      });

      if (serializedCredentialForm.success) {
        const response = await createLink({
          env,
          identifier,
          payload: serializedCredentialForm.data,
        });
        if (response.success) {
          setLinkID({ data: response.data.id, status: "successful" });
          setStep("summary");

          void messageAPI.success("Credential link created");
        } else {
          setLinkID({ error: null, status: "failed" });

          void messageAPI.error(response.error.message);
        }
      } else {
        notifyParseError(serializedCredentialForm.error);
      }
      setIsLoading(false);
    }
  };

  const issueCredential = async ({
    apiSchema,
    credentialIssuance,
    jsonSchema,
  }: {
    apiSchema: ApiSchema;
    credentialIssuance: CredentialDirectIssuance;
    jsonSchema: JsonSchema;
  }) => {
    const credentialSubjectAttribute = extractCredentialSubjectAttribute(jsonSchema);

    if (credentialSubjectAttribute) {
      setIsLoading(true);
      const serializedCredentialForm = serializeCredentialIssuance({
        attribute: credentialSubjectAttribute,
        credentialSchema: apiSchema.url,
        issueCredential: credentialIssuance,
        type: apiSchema.type,
      });

      if (serializedCredentialForm.success) {
        const response = await createCredential({
          env,
          identifier,
          payload: serializedCredentialForm.data,
        });
        if (response.success) {
          navigate(
            generatePath(ROUTES.credentials.path, {
              tabID: CREDENTIALS_TABS[0].tabID,
            })
          );

          if (credentialIssuance.mtProof) {
            void notifyChange("credential");
          }

          void messageAPI.success("Credential issued");
        } else {
          void messageAPI.error(response.error.message);
        }
      } else {
        notifyParseError(serializedCredentialForm.error);
      }

      setIsLoading(false);
    }
  };

  useEffect(() => {
    if (schemaID) {
      setCredentialFormInput((currentCredentialFormInput) => ({
        ...currentCredentialFormInput,
        issueCredential: { ...defaultCredentialFormInput.issueCredential, schemaID },
      }));
    }
  }, [schemaID]);

  return (
    <>
      {messageContext}

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
                  initialValues={credentialFormInput.issuanceMethod}
                  onChangeDid={onChangeDid}
                  onSubmit={(values) => {
                    setCredentialFormInput({ ...credentialFormInput, issuanceMethod: values });
                    setStep("issueCredential");
                  }}
                />
              );
            }

            case "issueCredential": {
              return (
                <Card className="issue-credential-card" title="Credential details">
                  <Space direction="vertical">
                    <IssueCredentialForm
                      initialValues={credentialFormInput.issueCredential}
                      isLoading={isLoading}
                      onBack={() => {
                        setStep("issuanceMethod");
                      }}
                      onSelectApiSchema={onSelectApiSchema}
                      onSubmit={({
                        apiSchema,
                        jsonSchema,
                        values,
                      }: {
                        apiSchema: ApiSchema;
                        jsonSchema: JsonSchema;
                        values: IssueCredentialFormData;
                      }) => {
                        const newCredentialFormInput: CredentialFormInput = {
                          ...credentialFormInput,
                          issueCredential: values,
                        };

                        setCredentialFormInput(newCredentialFormInput);

                        const parsedForm = credentialFormParser.safeParse(newCredentialFormInput);

                        if (parsedForm.success) {
                          if (parsedForm.data.type === "credentialLink") {
                            void createCredentialLink({
                              credentialLinkIssuance: parsedForm.data,
                              jsonSchema,
                            });
                          } else {
                            void issueCredential({
                              apiSchema,
                              credentialIssuance: parsedForm.data,
                              jsonSchema,
                            });
                          }
                        } else {
                          notifyParseError(parsedForm.error);
                        }
                      }}
                      type={credentialFormInput.issuanceMethod.type}
                    />
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
    </>
  );
}
