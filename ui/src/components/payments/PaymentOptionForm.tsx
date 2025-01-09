import { Button, Card, Divider, Flex, Form, Input, Modal, Select, Tag } from "antd";
import { useCallback, useEffect, useState } from "react";
import { getKeys } from "src/adapters/api/keys";
import { getPaymentConfigurations } from "src/adapters/api/payments";
import { PaymentConfigFormData, PaymentOptionFormData } from "src/adapters/parsers/view";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { AppError, Key, KeyType, PaymentConfigurations } from "src/domain";
import {
  AsyncTask,
  hasAsyncTaskFailed,
  isAsyncTaskDataAvailable,
  isAsyncTaskStarting,
} from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { SAVE, VALUE_REQUIRED } from "src/utils/constants";

function ConfigForm({
  keys,
  onCancel,
  onSubmit,
  open,
  paymentConfigurations,
}: {
  keys: Key[];
  onCancel: () => void;
  onSubmit: (formValues: PaymentConfigFormData) => void;
  open: boolean;
  paymentConfigurations: PaymentConfigurations;
}) {
  const [form] = Form.useForm<PaymentConfigFormData>();

  const handleAddConfig = () => {
    void form.validateFields().then((values) => {
      onSubmit(values);
      form.resetFields();
      onCancel();
    });
  };

  return (
    <Modal
      cancelText="Cancel"
      centered
      okText="Add"
      onCancel={onCancel}
      onOk={handleAddConfig}
      open={open}
      title="Add Configuration"
    >
      <Form form={form} layout="vertical">
        <Form.Item
          label="Payment option"
          name="paymentOptionID"
          rules={[{ message: VALUE_REQUIRED, required: true }]}
        >
          <Select className="full-width" placeholder="Choose a payment option">
            {Object.entries(paymentConfigurations).map(([optionID, optionDetail]) => (
              <Select.Option key={optionID} value={optionID}>
                {optionDetail.PaymentOption.Name}
              </Select.Option>
            ))}
          </Select>
        </Form.Item>

        <Form.Item
          label="Amount"
          name="amount"
          rules={[
            { message: VALUE_REQUIRED, required: true },
            {
              message: "Please enter a valid positive integer",
              validator: (_, value: string) => {
                return !value || /^[1-9]\d*$/.test(value) ? Promise.resolve() : Promise.reject();
              },
            },
          ]}
        >
          <Input placeholder="Enter amount" />
        </Form.Item>

        <Form.Item
          label="Recipient"
          name="recipient"
          rules={[{ message: VALUE_REQUIRED, required: true }]}
        >
          <Input placeholder="Enter recipient" />
        </Form.Item>

        <Form.Item
          label="Signin key"
          name="signingKeyID"
          rules={[{ message: VALUE_REQUIRED, required: true }]}
        >
          <Select className="full-width" placeholder="Choose a signin key">
            {keys.map(({ id, name }) => (
              <Select.Option key={id} value={id}>
                {name}
              </Select.Option>
            ))}
          </Select>
        </Form.Item>
        <Divider />
      </Form>
    </Modal>
  );
}

export function PaymentOptionForm({
  initialValies,
  onSubmit,
}: {
  initialValies: PaymentOptionFormData;
  onSubmit: (values: PaymentOptionFormData) => void;
}) {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();
  const [form] = Form.useForm<PaymentOptionFormData>();
  const configs = Form.useWatch<PaymentOptionFormData["paymentOptions"]>("paymentOptions", form);

  const [showConfigForm, setShowConfigForm] = useState(false);
  const [paymentConfigurations, setPaymentConfigurations] = useState<
    AsyncTask<PaymentConfigurations, AppError>
  >({
    status: "pending",
  });
  const [keys, setKeys] = useState<AsyncTask<Key[], AppError>>({
    status: "pending",
  });

  const fetchPaymentConfigurations = useCallback(
    async (signal?: AbortSignal) => {
      setPaymentConfigurations((previousConfigurations) =>
        isAsyncTaskDataAvailable(previousConfigurations)
          ? { data: previousConfigurations.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await getPaymentConfigurations({
        env,
        signal,
      });
      if (response.success) {
        setPaymentConfigurations({
          data: response.data,
          status: "successful",
        });
      } else {
        if (!isAbortedError(response.error)) {
          setPaymentConfigurations({ error: response.error, status: "failed" });
        }
      }
    },
    [env]
  );

  const fetchKeys = useCallback(
    async (signal?: AbortSignal) => {
      setKeys((previousKeys) =>
        isAsyncTaskDataAvailable(previousKeys)
          ? { data: previousKeys.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await getKeys({
        env,
        identifier,
        params: {
          type: KeyType.secp256k1,
        },
        signal,
      });
      if (response.success) {
        setKeys({
          data: response.data.items.successful,
          status: "successful",
        });
      } else {
        if (!isAbortedError(response.error)) {
          setKeys({ error: response.error, status: "failed" });
        }
      }
    },
    [env, identifier]
  );

  const handleAddConfig = (values: PaymentConfigFormData) => {
    const formValues = form.getFieldsValue();
    form.setFieldsValue({ ...formValues, paymentOptions: [...formValues.paymentOptions, values] });
  };

  const handleDeleteConfig = (index: number) => {
    const formValues = form.getFieldsValue();
    const updatedConfigs = formValues.paymentOptions.filter((_, idx) => index !== idx);

    form.setFieldsValue({
      ...formValues,
      paymentOptions: updatedConfigs,
    });
  };

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchKeys);

    return aborter;
  }, [fetchKeys]);

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchPaymentConfigurations);

    return aborter;
  }, [fetchPaymentConfigurations]);

  return (
    <>
      {(() => {
        if (hasAsyncTaskFailed(keys) || hasAsyncTaskFailed(paymentConfigurations)) {
          return (
            <Card className="centered">
              {hasAsyncTaskFailed(keys) && <ErrorResult error={keys.error.message} />};
              {hasAsyncTaskFailed(paymentConfigurations) && (
                <ErrorResult error={paymentConfigurations.error.message} />
              )}
              ;
            </Card>
          );
        } else if (isAsyncTaskStarting(keys) || isAsyncTaskStarting(paymentConfigurations)) {
          return (
            <Card className="centered">
              <LoadingResult />
            </Card>
          );
        } else {
          return (
            <>
              <Form form={form} initialValues={initialValies} layout="vertical" onFinish={onSubmit}>
                <Form.Item
                  label="Name"
                  name="name"
                  rules={[{ message: VALUE_REQUIRED, required: true }]}
                >
                  <Input placeholder="Enter name" />
                </Form.Item>

                <Form.Item
                  label="Description"
                  name="description"
                  rules={[{ message: VALUE_REQUIRED, required: true }]}
                >
                  <Input placeholder="Enter description" />
                </Form.Item>

                <Form.Item
                  label="Choose Configs"
                  name="paymentOptions"
                  rules={[
                    {
                      message: "Please add at least one config",
                      validator: (_, value: []) =>
                        value.length > 0 ? Promise.resolve() : Promise.reject(),
                    },
                  ]}
                >
                  <Flex vertical>
                    {configs && (
                      <Flex>
                        {configs.map((config, index) => {
                          const option = Object.entries(paymentConfigurations.data).find(
                            ([key]) => key === config.paymentOptionID
                          );

                          if (option) {
                            const [
                              ,
                              {
                                PaymentOption: { Name },
                              },
                            ] = option;

                            return (
                              <Tag
                                closable
                                key={`${Name}/${index}`}
                                onClose={() => handleDeleteConfig(index)}
                                style={{ margin: 8 }}
                              >
                                {Name}
                              </Tag>
                            );
                          }
                          return;
                        })}
                      </Flex>
                    )}

                    <Flex justify="center" style={{ margin: 8 }}>
                      <Button onClick={() => setShowConfigForm(true)}>Add config</Button>
                    </Flex>
                  </Flex>
                </Form.Item>

                <Divider />

                <Flex justify="flex-end">
                  <Button htmlType="submit" type="primary">
                    {SAVE}
                  </Button>
                </Flex>
              </Form>
              <ConfigForm
                keys={keys.data}
                onCancel={() => setShowConfigForm(false)}
                onSubmit={handleAddConfig}
                open={showConfigForm}
                paymentConfigurations={paymentConfigurations.data}
              />
            </>
          );
        }
      })()}
    </>
  );
}
