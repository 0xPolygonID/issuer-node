import { App, Button, Card, Col, Divider, Form, Input, Row, Select, Typography } from "antd";
import { useCallback, useEffect, useState } from "react";

import { getSupportedNetwork } from "src/adapters/api/identities";
import { IdentityFormData, identityFormDataParser } from "src/adapters/parsers/view";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { useEnvContext } from "src/contexts/Env";
import { AppError, IdentityType, Method, SupportedNetwork } from "src/domain";
import {
  AsyncTask,
  hasAsyncTaskFailed,
  isAsyncTaskDataAvailable,
  isAsyncTaskStarting,
} from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { VALUE_REQUIRED } from "src/utils/constants";
import { buildAppError } from "src/utils/error";

const initialValues: IdentityFormData = {
  blockchain: "",
  credentialStatusType: "",
  displayName: "",
  method: Method.iden3,
  network: "",
  type: IdentityType.BJJ,
};

function getNetworkFormValues(
  supportedNetworks: SupportedNetwork[],
  blockchainName?: string,
  networkName?: string
): Partial<IdentityFormData> | null {
  const foundNetwork = supportedNetworks.find(({ blockchain }) => blockchain === blockchainName);
  const selectedNetwork = foundNetwork ?? supportedNetworks[0];

  if (!selectedNetwork) {
    return null;
  }

  const { blockchain, networks } = selectedNetwork;
  const network = networks.find(({ name }) => name === networkName) || networks[0];
  const credentialStatusType = network.rhsMode[0];

  return {
    blockchain,
    credentialStatusType,
    network: network.name,
  };
}

export function IdentityForm({
  onSubmit,
  submitBtnText,
}: {
  onSubmit: (formValues: IdentityFormData) => void;
  submitBtnText: string;
}) {
  const env = useEnvContext();
  const [form] = Form.useForm<IdentityFormData>();
  const { message } = App.useApp();

  const [formData, setFormData] = useState<IdentityFormData>(initialValues);

  const [supportedNetworks, setSupportedNetworks] = useState<
    AsyncTask<SupportedNetwork[], AppError>
  >({
    status: "pending",
  });

  const fetchNetworks = useCallback(
    async (signal: AbortSignal) => {
      setSupportedNetworks((previousState) =>
        isAsyncTaskDataAvailable(previousState)
          ? { data: previousState.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await getSupportedNetwork({
        env,
        signal,
      });

      if (response.success) {
        if (response.data.failed.length) {
          void message.error(
            response.data.failed.map((error) => buildAppError(error).message).join("\n")
          );
        }

        setSupportedNetworks({ data: response.data.successful, status: "successful" });
        setFormData((prevFormData) => {
          if (response.data.successful.length) {
            return {
              ...prevFormData,
              ...getNetworkFormValues(response.data.successful),
            };
          }

          return prevFormData;
        });
      } else {
        if (!isAbortedError(response.error)) {
          setSupportedNetworks({ error: response.error, status: "failed" });
          void message.error(response.error.message);
        }
      }
    },
    [env, message]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchNetworks);
    return aborter;
  }, [fetchNetworks]);

  if (
    hasAsyncTaskFailed(supportedNetworks) ||
    (isAsyncTaskDataAvailable(supportedNetworks) && !supportedNetworks.data.length)
  ) {
    const errorMessage = hasAsyncTaskFailed(supportedNetworks)
      ? supportedNetworks.error.message
      : "";

    return (
      <Card className="centered">
        <ErrorResult
          error={[
            "An error occurred while downloading the supported networks from the API:",
            errorMessage,
          ].join("\n")}
        />
      </Card>
    );
  } else if (isAsyncTaskStarting(supportedNetworks)) {
    return (
      <Card className="centered">
        <LoadingResult />
      </Card>
    );
  } else {
    const blockchainOptions = supportedNetworks.data.map(({ blockchain }) => blockchain);

    const networkOptions = supportedNetworks.data
      .find(({ blockchain }) => blockchain === formData.blockchain)
      ?.networks.map(({ name }) => name);

    const credentialStatusOptions = supportedNetworks.data
      .find(({ blockchain }) => blockchain === formData.blockchain)
      ?.networks.find(({ name }) => name === formData.network)?.rhsMode;

    return (
      blockchainOptions.length &&
      networkOptions?.length &&
      credentialStatusOptions?.length && (
        <Form
          form={form}
          initialValues={formData}
          layout="vertical"
          onFinish={onSubmit}
          onValuesChange={(changedValue: Partial<IdentityFormData>, allValues) => {
            let updatedFormData = { ...allValues };

            if (changedValue.blockchain || changedValue.network) {
              updatedFormData = {
                ...allValues,
                ...getNetworkFormValues(
                  supportedNetworks.data,
                  allValues.blockchain,
                  allValues.network
                ),
              };
            }

            const parsedIdentityFormData = identityFormDataParser.safeParse(updatedFormData);

            if (parsedIdentityFormData.success) {
              setFormData(parsedIdentityFormData.data);
              form.setFieldsValue(parsedIdentityFormData.data);
            }
          }}
        >
          <Form.Item>
            <Form.Item
              label="Identity name"
              name="displayName"
              rules={[{ message: VALUE_REQUIRED, required: true }]}
            >
              <Input placeholder="Enter name" />
            </Form.Item>
            <Typography.Text type="secondary">
              Give your identity a descriptive name, e.g. “Age credential testing”. This name is
              only seen locally.
            </Typography.Text>
          </Form.Item>

          <Form.Item>
            <Form.Item
              label="Method"
              name="method"
              rules={[{ message: VALUE_REQUIRED, required: true }]}
            >
              <Select
                className="full-width"
                disabled={Object.values(Method).length < 2}
                placeholder="Method"
              >
                {Object.values(Method).map((method) => (
                  <Select.Option key={method} value={method}>
                    {method}
                  </Select.Option>
                ))}
              </Select>
            </Form.Item>
            <Typography.Text type="secondary">
              The protocol or system used to create, resolve, and manage the DID.
            </Typography.Text>
          </Form.Item>

          <Form.Item
            label="Blockchain"
            name="blockchain"
            rules={[{ message: VALUE_REQUIRED, required: true }]}
          >
            <Select className="full-width" placeholder="Type">
              {blockchainOptions.map((blockchain) => (
                <Select.Option key={blockchain} value={blockchain}>
                  {blockchain}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            label="Network"
            name="network"
            rules={[{ message: VALUE_REQUIRED, required: true }]}
          >
            <Select className="full-width" placeholder="Network">
              {networkOptions.map((network) => (
                <Select.Option key={network} value={network}>
                  {network}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item label="Type" name="type" rules={[{ message: VALUE_REQUIRED, required: true }]}>
            <Select className="full-width" placeholder="Type">
              {Object.values(IdentityType).map((type) => (
                <Select.Option key={type} value={type}>
                  {type}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item>
            <Form.Item
              label="Credential Status"
              name="credentialStatusType"
              rules={[{ message: VALUE_REQUIRED, required: true }]}
            >
              <Select className="full-width" placeholder="Credential Status">
                {Object.values(credentialStatusOptions).map((credentialStatus) => (
                  <Select.Option key={credentialStatus} value={credentialStatus}>
                    {credentialStatus}
                  </Select.Option>
                ))}
              </Select>
            </Form.Item>
            <Typography.Text type="secondary">
              Identity signing key&apos;s credential status is checked by clients to generate
              zero-knowledge proofs using signed credentials.
            </Typography.Text>
          </Form.Item>

          <>
            <Divider />
            <Row gutter={[8, 8]} justify="end">
              <Col>
                <Button htmlType="submit" type="primary">
                  {submitBtnText}
                </Button>
              </Col>
            </Row>
          </>
        </Form>
      )
    );
  }
}
