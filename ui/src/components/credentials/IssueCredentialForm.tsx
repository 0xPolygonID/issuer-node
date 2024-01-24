import Ajv, { ErrorObject } from "ajv";
import Ajv2020 from "ajv/dist/2020";
import addFormats from "ajv-formats";
import applyDraft2019Formats from "ajv-formats-draft2019";
import {
  Button,
  Checkbox,
  Col,
  DatePicker,
  Divider,
  Form,
  Input,
  Row,
  Select,
  Space,
  Typography,
  message,
} from "antd";
import { Store } from "antd/es/form/interface";
import dayjs from "dayjs";
import { useCallback, useEffect, useState } from "react";
import { z } from "zod";

import { getApiSchemas } from "src/adapters/api/schemas";
import { getJsonSchemaFromUrl } from "src/adapters/jsonSchemas";
import {dayjsInstanceParser, IssueCredentialFormData, serializeSchemaForm} from "src/adapters/parsers/forms";
import IconBack from "src/assets/icons/arrow-narrow-left.svg?react";
import IconRight from "src/assets/icons/arrow-narrow-right.svg?react";
import IconCheckMark from "src/assets/icons/check.svg?react";
import IconCopy from "src/assets/icons/copy-01.svg?react";
import { InputErrors, ObjectAttributeForm } from "src/components/credentials/ObjectAttributeForm";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { useEnvContext } from "src/contexts/Env";
import { ApiSchema, AppError, Attribute, JsonSchema, ObjectAttribute } from "src/domain";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import {
  ISSUE_CREDENTIAL_DIRECT,
  ISSUE_CREDENTIAL_LINK,
  SCHEMA_HASH,
  SCHEMA_TYPE,
  URL_FIELD_ERROR_MESSAGE,
  VALUE_REQUIRED,
} from "src/utils/constants";
import { buildAppError, jsonSchemaErrorToString, notifyError } from "src/utils/error";
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
  initialValues,
  isLoading,
  onBack,
  onSelectApiSchema,
  onSubmit,
  type,
}: {
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
  const [form] = Form.useForm<IssueCredentialFormData>();

  const [messageAPI, messageContext] = message.useMessage();

  const [apiSchema, setApiSchema] = useState<ApiSchema>();
  const [apiSchemas, setApiSchemas] = useState<AsyncTask<ApiSchema[], undefined>>({
    status: "pending",
  });

  const [jsonSchema, setJsonSchema] = useState<AsyncTask<JsonSchema, AppError>>({
    status: "pending",
  });

  const [inputErrors, setInputErrors] = useState<InputErrors>();

  const [refreshServiceChecked, setRefreshServiceChecked] = useState(false);

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
          ajv.addVocabulary(["$metadata"]);
          applyDraft2019Formats(ajv);

          const validate = ajv.compile({
            properties,
            required,
            type,
          });
          const valid = validate(serializedSchemaForm.data);

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
          notifyError(buildAppError(error));
        }
      } else {
        notifyError(buildAppError(serializedSchemaForm.error));
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
          const initialValuesWithSchemaValues: Store = credentialSubject
            ? {
                ...initialValues,
                credentialSubject: computeFormObjectInitialValues(
                  credentialSubject,
                  initialValues.credentialSubject || {}
                ),
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
    [computeFormObjectInitialValues, env, form, initialValues]
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
          void messageAPI.error(response.error.message);
        }
      }
    },
    [env, fetchJsonSchema, initialValues.schemaID, messageAPI]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchSchemas);

    return aborter;
  }, [fetchSchemas]);

  return (
    <>
      {messageContext}
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
          }
        }}
        onValuesChange={(_, values: IssueCredentialFormData) => {
          const jsonSchemaData = isAsyncTaskDataAvailable(jsonSchema) ? jsonSchema.data : undefined;
          const credentialSubjectAttributeWithoutId =
            jsonSchemaData && extractCredentialSubjectAttributeWithoutId(jsonSchemaData);
          values.credentialSubject &&
            credentialSubjectAttributeWithoutId &&
            isFormValid(values.credentialSubject, credentialSubjectAttributeWithoutId);
        }}
      >
        <Form.Item
          label="Select schema type"
          name="schemaID"
          rules={[{ message: VALUE_REQUIRED, required: true }]}
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

        {apiSchema && (
          <>
            <Form.Item>
              <Space direction="vertical">
                <Row justify="space-between">
                  <Typography.Text type="secondary">{SCHEMA_HASH}</Typography.Text>

                  <Typography.Text
                    copyable={{ icon: [<IconCopy key={0} />, <IconCheckMark key={1} />] }}
                  >
                    {apiSchema.hash}
                  </Typography.Text>
                </Row>
              </Space>
            </Form.Item>

            <Divider />

            <Typography.Paragraph>{apiSchema.type}</Typography.Paragraph>

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
                      {jsonSchema.data.schema.description && (
                        <Typography.Paragraph type="secondary">
                          {jsonSchema.data.schema.description}
                        </Typography.Paragraph>
                      )}

                      <Space direction="vertical" size="large">
                        <ObjectAttributeForm
                          attributes={credentialSubjectAttributeWithoutId.schema.attributes}
                          inputErrors={inputErrors}
                        />

                        <Form.Item
                          label="Proof type"
                          name="proofTypes"
                          rules={[{ message: VALUE_REQUIRED, required: true }]}
                        >
                          <Checkbox.Group>
                            <Space direction="vertical">
                              <Checkbox value="SIG">
                                <Typography.Text>Signature-based (SIG)</Typography.Text>

                                <Typography.Text type="secondary">
                                  Credential signed by the issuer using a BJJ private key.
                                </Typography.Text>
                              </Checkbox>

                              <Checkbox value="MTP">
                                <Typography.Text>Merkle Tree Proof (MTP)</Typography.Text>

                                <Typography.Text type="secondary">
                                  Credential will be added to the issuer&apos;s state tree. The
                                  state transition involves an on-chain transaction and gas fees.
                                </Typography.Text>
                              </Checkbox>
                            </Space>
                          </Checkbox.Group>
                        </Form.Item>
                      </Space>
                      <Form.Item label="Refresh Service">
                        <Space direction="vertical">
                          <Form.Item
                            name={["refreshService", "enabled"]}
                            noStyle
                            valuePropName="checked"
                          >
                            <Checkbox
                              checked={refreshServiceChecked}
                              onChange={() => {
                                setRefreshServiceChecked(!refreshServiceChecked);
                              }}
                            >
                              Enable
                            </Checkbox>
                          </Form.Item>
                          <Form.Item
                            name={["refreshService", "url"]}
                            rules={[
                              {
                                message: URL_FIELD_ERROR_MESSAGE,
                                validator: (_, value) =>
                                  refreshServiceChecked
                                    ? z.string().url().parseAsync(value)
                                    : Promise.resolve(true),
                              }
                             ]}
                           >
                            <Input
                              disabled={!refreshServiceChecked}
                              placeholder="Valid URL of the credential refresh service"
                            />
                          </Form.Item>
                        </Space>
                      </Form.Item>
                      <Form.Item
                        label="Credential expiration date"
                        name="credentialExpiration"
                        rules={[
                          {
                            message: 'Credential expiration must set when refresh service is enabled.',
                            validator: (_, value) =>
                              refreshServiceChecked
                                ? dayjsInstanceParser.parseAsync(value)
                                : Promise.resolve(true),
                          }
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
    </>
  );
}
