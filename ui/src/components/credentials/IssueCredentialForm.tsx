import { Ajv, ErrorObject } from "ajv";
import { Ajv2020 } from "ajv/dist/2020";
import addFormats from "ajv-formats";
import applyDraft2019Formats from "ajv-formats-draft2019";
import {
  App,
  Button,
  Checkbox,
  Col,
  DatePicker,
  Divider,
  Flex,
  Form,
  Input,
  Row,
  Select,
  Space,
  Typography,
} from "antd";
import { Store } from "antd/es/form/interface";
import dayjs from "dayjs";
import { useCallback, useEffect, useState } from "react";
import { generatePath } from "react-router-dom";
import { z } from "zod";

import { getDisplayMethods } from "src/adapters/api/display-method";
import { getSupportedBlockchains } from "src/adapters/api/identities";
import { getApiSchemas } from "src/adapters/api/schemas";
import { getJsonSchemaFromUrl } from "src/adapters/jsonSchemas";
import { buildAppError, jsonSchemaErrorToString, notifyError } from "src/adapters/parsers";
import {
  IssueCredentialFormData,
  dayjsInstanceParser,
  serializeSchemaForm,
} from "src/adapters/parsers/view";
import IconBack from "src/assets/icons/arrow-narrow-left.svg?react";
import IconRight from "src/assets/icons/arrow-narrow-right.svg?react";
import IconLink from "src/assets/icons/link-external-01.svg?react";
import { iso31661Countries } from "src/assets/iso31661Countries";
import { InputErrors, ObjectAttributeForm } from "src/components/credentials/ObjectAttributeForm";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import {
  ApiSchema,
  AppError,
  Attribute,
  CredentialStatusType,
  DisplayMethod,
  JsonSchema,
  ObjectAttribute,
  ProofType,
} from "src/domain";
import { ROUTES } from "src/routes";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import {
  ISSUE_CREDENTIAL_DIRECT,
  ISSUE_CREDENTIAL_LINK,
  SCHEMA_TYPE,
  URL_FIELD_ERROR_MESSAGE,
  VALUE_REQUIRED,
} from "src/utils/constants";
import {
  extractCredentialSubjectAttributeWithoutId,
  makeAttributeOptional,
} from "src/utils/jsonSchemas";

function addErrorToPath(inputErrors: InputErrors, path: string[], error: string): InputErrors {
  const key = path[0];
  if (path.length > 1) {
    const value = (key && inputErrors[key]) || {};
    return key
      ? {
          ...inputErrors,
          [key]: addErrorToPath(
            typeof value === "string" ? {} : value,
            path.slice(1, path.length),
            error
          ),
        }
      : inputErrors;
  } else {
    return key ? { ...inputErrors, [key]: error } : inputErrors;
  }
}

export function IssueCredentialForm({
  did,
  initialValues,
  isLoading,
  onBack,
  onSelectApiSchema,
  onSubmit,
  type,
}: {
  did?: string;
  initialValues: IssueCredentialFormData;
  isLoading: boolean;
  onBack: () => void;
  onSelectApiSchema: (apiSchema: ApiSchema) => void;
  onSubmit: (params: {
    apiSchema: ApiSchema;
    jsonSchema: JsonSchema;
    values: IssueCredentialFormData;
  }) => void;
  type: "directIssue" | "credentialLink";
}) {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();
  const [form] = Form.useForm<IssueCredentialFormData>();

  const { message } = App.useApp();

  const [apiSchema, setApiSchema] = useState<ApiSchema>();
  const [apiSchemas, setApiSchemas] = useState<AsyncTask<ApiSchema[], undefined>>({
    status: "pending",
  });

  const [jsonSchema, setJsonSchema] = useState<AsyncTask<JsonSchema, AppError>>({
    status: "pending",
  });

  const [displayMethods, setDisplayMethods] = useState<AsyncTask<DisplayMethod[], AppError>>({
    status: "pending",
  });

  const [credentialStatusTypes, setCredentialStatusTypes] = useState<
    AsyncTask<CredentialStatusType[], AppError>
  >({
    status: "pending",
  });

  const [inputErrors, setInputErrors] = useState<InputErrors>();

  const displayMethodChecked =
    Form.useWatch<IssueCredentialFormData["displayMethod"]["enabled"]>(
      ["displayMethod", "enabled"],
      form
    ) || initialValues.displayMethod.enabled;

  const refreshServiceChecked =
    Form.useWatch<IssueCredentialFormData["refreshService"]["enabled"]>(
      ["refreshService", "enabled"],
      form
    ) || initialValues.refreshService.enabled;

  const isPositiveBigInt = (x: string) => {
    try {
      return BigInt(x).toString() === x && BigInt(x) > 0;
    } catch {
      return false;
    }
  };

  const isNonNegativeBigInt = (x: string) => {
    try {
      return BigInt(x).toString() === x && BigInt(x) >= 0;
    } catch {
      return false;
    }
  };

  const isCountryCode = (code: number) => {
    return iso31661Countries.some((country) => country.code === code);
  };

  function isFormValid(value: Record<string, unknown>, objectAttribute: ObjectAttribute): boolean {
    if (isAsyncTaskDataAvailable(jsonSchema)) {
      const serializedSchemaForm = serializeSchemaForm({
        attribute: makeAttributeOptional(objectAttribute),
        value,
      });

      if (serializedSchemaForm.success) {
        const { properties, required, type } = objectAttribute.schema;

        try {
          const ajv =
            jsonSchema.data.jsonSchemaProps.$schema ===
            "https://json-schema.org/draft/2020-12/schema"
              ? new Ajv2020({ allErrors: true })
              : new Ajv({ allErrors: true });
          addFormats(ajv);
          ajv.addFormat("positive-integer", {
            type: "string",
            validate: isPositiveBigInt,
          });
          ajv.addFormat("positive-integer-eth-address", {
            type: "string",
            validate: isPositiveBigInt,
          });
          ajv.addFormat("non-negative-integer", {
            type: "string",
            validate: isNonNegativeBigInt,
          });
          ajv.addFormat("iso-3166-1-numeric", {
            type: "number",
            validate: isCountryCode,
          });
          ajv.addVocabulary(["$metadata"]);
          applyDraft2019Formats(ajv);

          const validate = ajv.compile({
            properties,
            required,
            type,
          });

          const valid = validate(serializedSchemaForm.data || {});

          if (valid) {
            setInputErrors(undefined);
            return true;
          } else if (validate.errors) {
            setInputErrors(
              validate.errors.reduce((acc: InputErrors, curr: ErrorObject): InputErrors => {
                if (curr.keyword === "required") {
                  // filtering out required errors since we manage these from the antd form
                  return acc;
                } else {
                  const errorMsg = curr.message
                    ? curr.message.charAt(0).toUpperCase() + curr.message.slice(1)
                    : "Unknown validation error";
                  const path = curr.instancePath
                    .split("/")
                    .filter((segment) => segment !== "/" && segment !== "");
                  return addErrorToPath(acc, path, errorMsg);
                }
              }, {})
            );
          }
        } catch (error) {
          void notifyError(buildAppError(error));
        }
      } else {
        void notifyError(buildAppError(serializedSchemaForm.error));
      }
    }
    return false;
  }

  const computeFormObjectInitialValues = useCallback(
    (
      objectAttribute: ObjectAttribute,
      initialValues: Record<string, unknown>
    ): Record<string, unknown> | undefined => {
      return objectAttribute.schema.attributes?.reduce(
        (acc: Record<string, unknown>, curr: Attribute): Record<string, unknown> => {
          switch (curr.type) {
            case "boolean": {
              const parsedConst = z.boolean().safeParse(curr.schema.const);
              const parsedDefault = z.boolean().safeParse(curr.schema.default);
              const constValue = parsedConst.success ? parsedConst.data : undefined;
              const defaultValue = parsedDefault.success ? parsedDefault.data : undefined;
              const value = constValue !== undefined ? constValue : defaultValue;
              return { ...acc, [curr.name]: value };
            }
            case "integer":
            case "number": {
              const parsedConst = z.number().safeParse(curr.schema.const);
              const parsedDefault = z.number().safeParse(curr.schema.default);
              const constValue = parsedConst.success ? parsedConst.data : undefined;
              const defaultValue = parsedDefault.success ? parsedDefault.data : undefined;
              const value = constValue !== undefined ? constValue : defaultValue;
              return { ...acc, [curr.name]: value };
            }
            case "string": {
              const parsedConst = z.string().safeParse(curr.schema.const);
              const parsedDefault = z.string().safeParse(curr.schema.default);
              const constValue = parsedConst.success ? parsedConst.data : undefined;
              const defaultValue = parsedDefault.success ? parsedDefault.data : undefined;
              const value = constValue !== undefined ? constValue : defaultValue;
              if (value === undefined) {
                return acc;
              }
              switch (curr.schema.format) {
                case "date":
                case "date-time": {
                  return { ...acc, [curr.name]: dayjs(value) };
                }
                case "time": {
                  return { ...acc, [curr.name]: dayjs(`1970-01-01T${value}`) };
                }
                default: {
                  return { ...acc, [curr.name]: value };
                }
              }
            }
            case "object": {
              const parsedRecord = z.record(z.unknown()).safeParse(initialValues[curr.name] || {});
              return parsedRecord.success
                ? {
                    ...acc,
                    [curr.name]: computeFormObjectInitialValues(curr, parsedRecord.data),
                  }
                : acc;
            }
            default: {
              return acc;
            }
          }
        },
        initialValues
      );
    },
    []
  );

  const fetchJsonSchema = useCallback(
    (schema: ApiSchema) => {
      setJsonSchema({ status: "loading" });
      void getJsonSchemaFromUrl({
        env,
        url: schema.url,
      }).then((response) => {
        if (response.success) {
          const [jsonSchema] = response.data;
          setJsonSchema({
            data: jsonSchema,
            status: "successful",
          });
          const credentialSubject = extractCredentialSubjectAttributeWithoutId(jsonSchema);
          const schemaDefaultDisplayMethod = isAsyncTaskDataAvailable(displayMethods)
            ? displayMethods.data.find(({ id }) => id === schema.displayMethodID)
            : undefined;

          const initialValuesWithSchemaValues: Store = credentialSubject
            ? {
                ...initialValues,
                credentialSubject: computeFormObjectInitialValues(
                  credentialSubject,
                  initialValues.credentialSubject || {}
                ),
                displayMethod: {
                  enabled: !!schemaDefaultDisplayMethod,
                  ...(schemaDefaultDisplayMethod
                    ? { type: schemaDefaultDisplayMethod.type, url: schemaDefaultDisplayMethod.url }
                    : { type: "", url: null }),
                },
              }
            : initialValues;
          form.setFieldsValue(initialValuesWithSchemaValues);
        } else {
          if (!isAbortedError(response.error)) {
            setJsonSchema({ error: response.error, status: "failed" });
          }
        }
      });
    },
    [computeFormObjectInitialValues, env, form, initialValues, displayMethods]
  );

  const fetchSchemas = useCallback(
    async (signal: AbortSignal) => {
      setApiSchemas((previousState) =>
        isAsyncTaskDataAvailable(previousState)
          ? { data: previousState.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await getApiSchemas({
        env,
        identifier,
        params: {},
        signal,
      });

      if (response.success) {
        setApiSchemas({ data: response.data.successful, status: "successful" });
        const selectedSchema =
          initialValues.schemaID !== undefined
            ? response.data.successful.find((schema) => schema.id === initialValues.schemaID)
            : undefined;

        if (selectedSchema) {
          setApiSchema(selectedSchema);
          fetchJsonSchema(selectedSchema);
        }
      } else {
        if (!isAbortedError(response.error)) {
          setApiSchemas({ error: undefined, status: "failed" });
          void message.error(response.error.message);
        }
      }
    },
    [env, fetchJsonSchema, initialValues.schemaID, message, identifier]
  );

  const fetchDisplayMethods = useCallback(
    async (signal?: AbortSignal) => {
      setDisplayMethods((previousDisplayMethods) =>
        isAsyncTaskDataAvailable(previousDisplayMethods)
          ? { data: previousDisplayMethods.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await getDisplayMethods({
        env,
        identifier,
        params: {},
        signal,
      });
      if (response.success) {
        setDisplayMethods({
          data: response.data.items.successful,
          status: "successful",
        });
      } else {
        if (!isAbortedError(response.error)) {
          setDisplayMethods({ error: response.error, status: "failed" });
        }
      }
    },
    [env, identifier]
  );

  const fetchBlockChains = useCallback(
    async (signal: AbortSignal) => {
      setCredentialStatusTypes((previousState) =>
        isAsyncTaskDataAvailable(previousState)
          ? { data: previousState.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await getSupportedBlockchains({
        env,
        signal,
      });

      if (response.success) {
        const [, , blockchain = "", network = ""] = identifier.split(":");
        const identityBlockchainNetworks =
          response.data.successful.find(({ name }) => name === blockchain)?.networks || [];
        const identityNetworkCredentialStatusTypes =
          identityBlockchainNetworks.find(({ name }) => name === network)?.credentialStatus || [];

        setCredentialStatusTypes({
          data: identityNetworkCredentialStatusTypes,
          status: "successful",
        });
      } else {
        if (!isAbortedError(response.error)) {
          setCredentialStatusTypes({ error: response.error, status: "failed" });
          void message.error(response.error.message);
        }
      }
    },
    [env, message, identifier]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchSchemas);

    return aborter;
  }, [fetchSchemas]);

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchDisplayMethods);

    return aborter;
  }, [fetchDisplayMethods]);

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchBlockChains);

    return aborter;
  }, [fetchBlockChains]);

  return (
    <Form
      form={form}
      initialValues={initialValues}
      layout="vertical"
      onFinish={(values: IssueCredentialFormData) => {
        const jsonSchemaData = isAsyncTaskDataAvailable(jsonSchema) ? jsonSchema.data : undefined;
        const credentialSubjectAttributeWithoutId =
          jsonSchemaData && extractCredentialSubjectAttributeWithoutId(jsonSchemaData);

        if (
          jsonSchemaData &&
          credentialSubjectAttributeWithoutId &&
          values.credentialSubject &&
          isFormValid(values.credentialSubject, credentialSubjectAttributeWithoutId) &&
          apiSchema
        ) {
          onSubmit({ apiSchema, jsonSchema: jsonSchemaData, values });
        } else {
          void message.error("Error validating the data against the schema");
        }
      }}
      onValuesChange={(
        updatedValue: Partial<IssueCredentialFormData>,
        values: IssueCredentialFormData
      ) => {
        if (updatedValue.displayMethod?.url && isAsyncTaskDataAvailable(displayMethods)) {
          const displayMethod = displayMethods.data.find(
            ({ url }) => url === updatedValue.displayMethod?.url
          );
          if (displayMethod) {
            form.setFieldValue("displayMethod", {
              ...values.displayMethod,
              type: displayMethod.type,
              url: displayMethod.url,
            });
          }
        }

        const jsonSchemaData = isAsyncTaskDataAvailable(jsonSchema) ? jsonSchema.data : undefined;
        const credentialSubjectAttributeWithoutId =
          jsonSchemaData && extractCredentialSubjectAttributeWithoutId(jsonSchemaData);
        values.credentialSubject &&
          credentialSubjectAttributeWithoutId &&
          isFormValid(values.credentialSubject, credentialSubjectAttributeWithoutId);
      }}
    >
      {did && apiSchema && (
        <>
          <Flex justify="space-between" vertical>
            <Typography.Text>Recipient identifier:</Typography.Text>
            <Typography.Text>{did}</Typography.Text>
          </Flex>
          <Divider />
        </>
      )}

      <Flex align="flex-end" gap={8} justify="space-between">
        <Form.Item
          label="Select schema type"
          name="schemaID"
          rules={[{ message: VALUE_REQUIRED, required: true }]}
          style={{ marginBottom: 0, width: "100%" }}
        >
          <Select
            className="full-width"
            loading={isAsyncTaskStarting(apiSchemas)}
            onChange={(id: string) => {
              const schema =
                isAsyncTaskDataAvailable(apiSchemas) &&
                apiSchemas.data.find((schema) => schema.id === id);
              if (schema) {
                onSelectApiSchema(schema);
                setApiSchema(schema);
                fetchJsonSchema(schema);
              }
            }}
            placeholder={SCHEMA_TYPE}
          >
            {isAsyncTaskDataAvailable(apiSchemas) &&
              apiSchemas.data.map(({ id, type }) => (
                <Select.Option key={id} value={id}>
                  {type}
                </Select.Option>
              ))}
          </Select>
        </Form.Item>

        <Button
          disabled={!apiSchema}
          href={generatePath(ROUTES.schemaDetails.path, {
            schemaID: apiSchema?.id || "",
          })}
          icon={<IconLink />}
          target="_blank"
        />
      </Flex>

      {apiSchema && (
        <>
          <Divider />

          {(() => {
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
                const credentialSubjectAttributeWithoutId =
                  extractCredentialSubjectAttributeWithoutId(jsonSchema.data);

                return credentialSubjectAttributeWithoutId?.schema.attributes ? (
                  <>
                    <Space direction="vertical" size="large" style={{ rowGap: 0 }}>
                      <ObjectAttributeForm
                        attributes={credentialSubjectAttributeWithoutId.schema.attributes}
                        inputErrors={inputErrors}
                      />

                      <Divider />

                      <Form.Item
                        label="Proof type"
                        name="proofTypes"
                        rules={[{ message: VALUE_REQUIRED, required: true }]}
                        style={{ marginBottom: 0 }}
                      >
                        <Checkbox.Group>
                          <Space direction="vertical">
                            <Checkbox value={ProofType.BJJSignature2021}>
                              <Typography.Text>
                                Signature-based ({ProofType.BJJSignature2021})
                              </Typography.Text>

                              <Typography.Text type="secondary">
                                Credential signed by the issuer using a BJJ private key.
                              </Typography.Text>
                            </Checkbox>

                            <Checkbox value={ProofType.Iden3SparseMerkleTreeProof}>
                              <Typography.Text>
                                Merkle Tree Proof ({ProofType.Iden3SparseMerkleTreeProof})
                              </Typography.Text>

                              <Typography.Text type="secondary">
                                Credential will be added to the issuer&apos;s state tree. The state
                                transition involves an on-chain transaction and gas fees.
                              </Typography.Text>
                            </Checkbox>
                          </Space>
                        </Checkbox.Group>
                      </Form.Item>
                    </Space>
                    <Divider />

                    <Form.Item
                      label="Revocation status"
                      name="credentialStatusType"
                      style={{ marginBottom: 0 }}
                    >
                      <Select
                        className="full-width"
                        loading={isAsyncTaskStarting(credentialStatusTypes)}
                        placeholder="Choose revocation status"
                      >
                        {isAsyncTaskDataAvailable(credentialStatusTypes) &&
                          credentialStatusTypes.data.map((status) => (
                            <Select.Option key={status} value={status}>
                              {status}
                            </Select.Option>
                          ))}
                      </Select>
                    </Form.Item>

                    <Divider />

                    <Form.Item style={{ marginBottom: 0 }}>
                      <Space direction="vertical">
                        <Form.Item
                          name={["refreshService", "enabled"]}
                          noStyle
                          valuePropName="checked"
                        >
                          <Checkbox checked={refreshServiceChecked}>
                            Refresh Service{" ("}
                            <Typography.Link
                              href="https://docs.privado.id/docs/category/refresh-service"
                              style={{
                                alignItems: "center",
                                display: "inline-flex",
                                flexWrap: "nowrap",
                                gap: 4,
                              }}
                              target="_blank"
                            >
                              see documentation <IconLink style={{ width: 14 }} />
                            </Typography.Link>
                            {") "}
                          </Checkbox>
                        </Form.Item>
                        <Form.Item
                          hidden={!refreshServiceChecked}
                          name={["refreshService", "url"]}
                          rules={[
                            {
                              message: URL_FIELD_ERROR_MESSAGE,
                              validator: (_, value) =>
                                refreshServiceChecked
                                  ? z.string().url().parseAsync(value)
                                  : Promise.resolve(true),
                            },
                          ]}
                        >
                          <Input placeholder="Valid URL of the credential refresh service" />
                        </Form.Item>
                      </Space>
                    </Form.Item>
                    <Form.Item>
                      <Space direction="vertical">
                        <Form.Item
                          name={["displayMethod", "enabled"]}
                          noStyle
                          valuePropName="checked"
                        >
                          <Checkbox checked={displayMethodChecked}>Display Method</Checkbox>
                        </Form.Item>
                        <Form.Item hidden={!displayMethodChecked} name={["displayMethod", "url"]}>
                          <Select
                            className="full-width"
                            loading={isAsyncTaskStarting(displayMethods)}
                            placeholder="Select display method"
                          >
                            {isAsyncTaskDataAvailable(displayMethods) &&
                              displayMethods.data.map(({ id, name, url }) => (
                                <Select.Option key={id} value={url}>
                                  {name}
                                </Select.Option>
                              ))}
                          </Select>
                        </Form.Item>
                        <Form.Item hidden name={["displayMethod", "type"]}>
                          <Input />
                        </Form.Item>
                      </Space>
                    </Form.Item>
                    <Form.Item
                      label="Credential expiration date"
                      name="credentialExpiration"
                      rules={[
                        {
                          message:
                            "Credential expiration must be set when the refresh service is enabled",
                          validator: (_, value) =>
                            refreshServiceChecked
                              ? dayjsInstanceParser.parseAsync(value)
                              : Promise.resolve(true),
                        },
                      ]}
                    >
                      <DatePicker
                        format="YYYY-MM-DD HH:mm:ss"
                        showTime={{ defaultValue: dayjs("23:59:59", "HH:mm:ss") }}
                      />
                    </Form.Item>
                  </>
                ) : (
                  <ErrorResult error="An error occurred while getting the credentialSubject attributes of the JSON Schema" />
                );
              }
            }
          })()}
        </>
      )}

      {jsonSchema.status !== "failed" && (
        <>
          <Divider />
          <Row gutter={[8, 8]} justify="end">
            <Col>
              <Button icon={<IconBack />} onClick={onBack} type="default">
                Previous step
              </Button>
            </Col>

            <Col>
              <Button htmlType="submit" loading={isLoading} type="primary">
                {type === "directIssue" ? ISSUE_CREDENTIAL_DIRECT : ISSUE_CREDENTIAL_LINK}
                {type === "credentialLink" && <IconRight />}
              </Button>
            </Col>
          </Row>
        </>
      )}
    </Form>
  );
}
