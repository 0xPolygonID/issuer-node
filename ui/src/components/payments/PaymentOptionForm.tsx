import { Button, Card, Divider, Flex, Form, Input, Modal, Select } from "antd";
import { useCallback, useEffect, useState } from "react";
import { getKeys } from "src/adapters/api/keys";
import { PaymentConfigFormData, PaymentOptionFormData } from "src/adapters/parsers/view";
import IconInfoCircle from "src/assets/icons/info-circle.svg?react";
import { PaymentConfigTable } from "src/components/payments/PaymentConfigTable";
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
  initialValues,
  keys,
  onCancel,
  onSubmit,
  open,
  paymentConfigurations,
}: {
  initialValues?: PaymentConfigFormData;
  keys: Key[];
  onCancel: () => void;
  onSubmit: (formValues: PaymentConfigFormData) => void;
  open: boolean;
  paymentConfigurations: PaymentConfigurations;
}) {
  const [form] = Form.useForm<PaymentConfigFormData>();

  const handleSubmit = () => {
    void form.validateFields().then((values) => {
      onSubmit(values);
      onCancel();
    });
  };

  return (
    <Modal
      cancelText="Cancel"
      centered
      destroyOnClose={true}
      okText={SAVE}
      onCancel={onCancel}
      onOk={handleSubmit}
      open={open}
      title="Add Configuration"
    >
      <Form form={form} initialValues={initialValues} layout="vertical" preserve={false}>
        <Form.Item
          label="Payment option"
          name="paymentOptionID"
          rules={[{ message: VALUE_REQUIRED, required: true }]}
        >
          <Select
            className="full-width"
            onChange={(value: string) => {
              const paymentOption = paymentConfigurations[value];
              form.setFieldsValue({
                decimals: paymentOption?.PaymentOption.Decimals,
              });
            }}
            placeholder="Choose a payment option"
          >
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
              message: "Please enter a valid positive number",
              validator: (_, value: string) => {
                return !value || /^\d*\.?\d+$/.test(value) ? Promise.resolve() : Promise.reject();
              },
            },
          ]}
          tooltip={{
            icon: <IconInfoCircle style={{ width: 14 }} />,
            placement: "right",
            title:
              "Enter the amount in native currency (e.g., ETH, POL) or tokens (e.g., USDT, USDC) in a standard readable format, not in smallest units.",
          }}
        >
          <Input placeholder="Enter amount" />
        </Form.Item>

        <Form.Item
          label="Recipient"
          name="recipient"
          rules={[{ message: VALUE_REQUIRED, required: true }]}
          tooltip={{
            icon: <IconInfoCircle style={{ width: 14 }} />,
            placement: "right",
            title: "Enter the recipient's EOA address allowed to withdraw funds.",
          }}
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

        <Form.Item hidden name="decimals">
          <Input type="number" />
        </Form.Item>

        <Divider />
      </Form>
    </Modal>
  );
}

export function PaymentOptionForm({
  initialValies,
  onSubmit,
  paymentConfigurations,
}: {
  initialValies: PaymentOptionFormData;
  onSubmit: (values: PaymentOptionFormData) => void;
  paymentConfigurations: PaymentConfigurations;
}) {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();
  const [form] = Form.useForm<PaymentOptionFormData>();
  const configs = Form.useWatch<PaymentOptionFormData["paymentOptions"]>("paymentOptions", form);

  const [showConfigForm, setShowConfigForm] = useState(false);
  const [editableConfig, setEditableConfig] = useState<{
    index: number;
    initialValues: PaymentConfigFormData;
  } | null>(null);

  const [keys, setKeys] = useState<AsyncTask<Key[], AppError>>({
    status: "pending",
  });

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

  const handleUpsertConfig = useCallback(
    (values: PaymentConfigFormData) => {
      const formValues = form.getFieldsValue();

      if (editableConfig) {
        form.setFieldsValue({
          ...formValues,
          paymentOptions: formValues.paymentOptions.map((paymentOption, idx) =>
            idx === editableConfig.index ? values : paymentOption
          ),
        });
      } else {
        form.setFieldsValue({
          ...formValues,
          paymentOptions: [...formValues.paymentOptions, values],
        });
      }
    },
    [editableConfig, form]
  );

  const handleDeleteConfig = useCallback(
    (index: number) => {
      const formValues = form.getFieldsValue();
      const updatedConfigs = formValues.paymentOptions.filter((_, idx) => index !== idx);

      form.setFieldsValue({
        ...formValues,
        paymentOptions: updatedConfigs,
      });
    },
    [form]
  );

  const handleEditConfig = useCallback(
    (index: number) => {
      const config = configs.find((_, idx) => idx === index);
      if (config) {
        setShowConfigForm(true);
        setEditableConfig({
          index,
          initialValues: config,
        });
      }
    },
    [configs]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchKeys);

    return aborter;
  }, [fetchKeys]);

  return (
    <>
      {(() => {
        if (hasAsyncTaskFailed(keys)) {
          return (
            <Card className="centered">
              <ErrorResult
                error={[
                  "An error occurred while downloading a keys from the API:",
                  keys.error.message,
                ].join("\n")}
              />
            </Card>
          );
        } else if (isAsyncTaskStarting(keys)) {
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
                    <PaymentConfigTable
                      configs={
                        configs
                          ? configs.map(({ paymentOptionID, ...other }) => ({
                              paymentOptionID: parseInt(paymentOptionID),
                              ...other,
                            }))
                          : []
                      }
                      onDelete={handleDeleteConfig}
                      onEdit={handleEditConfig}
                      showTitle={false}
                    />

                    <Flex justify="center" style={{ marginTop: 16 }}>
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
                initialValues={editableConfig?.initialValues}
                keys={keys.data}
                onCancel={() => {
                  setEditableConfig(null);
                  setShowConfigForm(false);
                }}
                onSubmit={handleUpsertConfig}
                open={showConfigForm}
                paymentConfigurations={paymentConfigurations}
              />
            </>
          );
        }
      })()}
    </>
  );
}
