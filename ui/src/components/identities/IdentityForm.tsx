import { App, Button, Col, Divider, Form, Input, Row, Select, Typography } from "antd";
import { useCallback, useEffect, useState } from "react";
import { credentialStatusTypeParser } from "src/adapters/api/credentials";

import { getSupportedBlockchains } from "src/adapters/api/identities";
import { IdentityFormData, identityFormDataParser } from "src/adapters/parsers/view";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { useEnvContext } from "src/contexts/Env";
import { AppError, Blockchain, CredentialStatusType, IdentityType, Method } from "src/domain";
import { AsyncTask, isAsyncTaskDataAvailable } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { VALUE_REQUIRED } from "src/utils/constants";
import { buildAppError, notifyParseError } from "src/utils/error";

const initialValues: IdentityFormData = {
  blockchain: "",
  credentialStatusType: CredentialStatusType["Iden3commRevocationStatusV1.0"],
  displayName: "",
  method: Method.iden3,
  network: "",
  type: IdentityType.BJJ,
};

function getNetworkFormValues(
  blockchains: Blockchain[],
  blockchainName?: string,
  networkName?: string
): Partial<IdentityFormData> | null {
  const foundBlockchain = blockchains.find(({ name }) => name === blockchainName);
  const selectedBlockchain = foundBlockchain ?? blockchains[0];

  if (!selectedBlockchain) {
    return null;
  }

  const { name, networks } = selectedBlockchain;
  const network = networks.find(({ name }) => name === networkName) || networks[0];
  const parsedCredentialStatusType = credentialStatusTypeParser.safeParse(network.rhsMode[0]);

  if (parsedCredentialStatusType.success) {
    return {
      blockchain: name,
      credentialStatusType: parsedCredentialStatusType.data,
      network: network.name,
    };
  } else {
    notifyParseError(parsedCredentialStatusType.error);
    return null;
  }
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

  const [blockchains, setBlockchains] = useState<AsyncTask<Blockchain[], AppError>>({
    status: "pending",
  });

  const fetchNetworks = useCallback(
    async (signal: AbortSignal) => {
      setBlockchains((previousState) =>
        isAsyncTaskDataAvailable(previousState)
          ? { data: previousState.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await getSupportedBlockchains({
        env,
        signal,
      });

      if (response.success) {
        if (response.data.failed.length) {
          void message.error(
            response.data.failed.map((error) => buildAppError(error).message).join("\n")
          );
        }

        setBlockchains({ data: response.data.successful, status: "successful" });
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
          setBlockchains({ error: response.error, status: "failed" });
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

  switch (blockchains.status) {
    case "pending":
    case "loading": {
      return <LoadingResult />;
    }
    case "failed": {
      return (
        <ErrorResult
          error={[
            "An error occurred while downloading the supported networks from the API:",
            blockchains.error.message,
          ].join("\n")}
        />
      );
    }
    case "reloading":
    case "successful": {
      const blockchainOptions = blockchains.data.map(({ name }) => name);

      const networkOptions = blockchains.data
        .find(({ name }) => name === formData.blockchain)
        ?.networks.map(({ name }) => name);

      const credentialStatusOptions = blockchains.data
        .find(({ name }) => name === formData.blockchain)
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
                    blockchains.data,
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

            <Form.Item
              label="Type"
              name="type"
              rules={[{ message: VALUE_REQUIRED, required: true }]}
            >
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
}
