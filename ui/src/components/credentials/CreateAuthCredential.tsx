import { App, Button, Card, Divider, Flex, Form, Select, Space } from "antd";
import { useCallback, useEffect, useState } from "react";
import { generatePath, useNavigate } from "react-router-dom";
import {
  CreateAuthCredential as CreateAuthCredentialType,
  createAuthCredential,
} from "src/adapters/api/credentials";
import { getSupportedBlockchains } from "src/adapters/api/identities";
import { getKeys } from "src/adapters/api/keys";
import { notifyErrors } from "src/adapters/parsers";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";

import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { AppError, CredentialStatusType, Key, KeyType } from "src/domain";
import { ROUTES } from "src/routes";
import {
  AsyncTask,
  hasAsyncTaskFailed,
  isAsyncTaskDataAvailable,
  isAsyncTaskStarting,
} from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { VALUE_REQUIRED } from "src/utils/constants";

export function CreateAuthCredential() {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();
  const [form] = Form.useForm<CreateAuthCredentialType>();
  const navigate = useNavigate();
  const { message } = App.useApp();

  const [keys, setKeys] = useState<AsyncTask<Key[], AppError>>({
    status: "pending",
  });

  const [credentialStatusTypes, setCredentialStatusTypes] = useState<
    AsyncTask<CredentialStatusType[], AppError>
  >({
    status: "pending",
  });

  const handleSubmit = (formValues: CreateAuthCredentialType) => {
    return void createAuthCredential({
      env,
      identifier,
      payload: formValues,
    }).then((response) => {
      if (response.success) {
        void message.success("Auth credential added successfully");
        navigate(generatePath(ROUTES.identityDetails.path, { identityID: identifier }));
      } else {
        void message.error(response.error.message);
      }
    });
  };

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
          type: KeyType.babyjubJub,
        },
        signal,
      });
      if (response.success) {
        setKeys({
          data: response.data.items.successful,
          status: "successful",
        });

        void notifyErrors(response.data.items.failed);
      } else {
        if (!isAbortedError(response.error)) {
          setKeys({ error: response.error, status: "failed" });
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
    const { aborter } = makeRequestAbortable(fetchKeys);

    return aborter;
  }, [fetchKeys]);

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchBlockChains);

    return aborter;
  }, [fetchBlockChains]);

  return (
    <SiderLayoutContent
      description="Create a new auth credential"
      showBackButton
      showDivider
      title="Add new auth credential"
    >
      {(() => {
        if (hasAsyncTaskFailed(keys) || hasAsyncTaskFailed(credentialStatusTypes)) {
          return (
            <Card className="centered">
              {hasAsyncTaskFailed(keys) && <ErrorResult error={keys.error.message} />};
              {hasAsyncTaskFailed(credentialStatusTypes) && (
                <ErrorResult error={credentialStatusTypes.error.message} />
              )}
              ;
            </Card>
          );
        } else if (isAsyncTaskStarting(keys) || isAsyncTaskStarting(credentialStatusTypes)) {
          return (
            <Card className="centered">
              <LoadingResult />
            </Card>
          );
        } else {
          return (
            <Card className="centered" title="Auth credential details">
              <Space direction="vertical" size="large">
                <Form form={form} layout="vertical" onFinish={handleSubmit}>
                  <Form.Item
                    label="Key"
                    name="keyID"
                    rules={[{ message: VALUE_REQUIRED, required: true }]}
                  >
                    <Select className="full-width" placeholder="Choose a key">
                      {keys.data.map(({ id, name }) => (
                        <Select.Option key={id} value={id}>
                          {name}
                        </Select.Option>
                      ))}
                    </Select>
                  </Form.Item>

                  <Form.Item
                    label="Credential status"
                    name="credentialStatusType"
                    rules={[{ message: VALUE_REQUIRED, required: true }]}
                  >
                    <Select className="full-width" placeholder="Choose a credential status type">
                      {credentialStatusTypes.data.map((type) => (
                        <Select.Option key={type} value={type}>
                          {type}
                        </Select.Option>
                      ))}
                    </Select>
                  </Form.Item>

                  <Divider />

                  <Flex justify="flex-end">
                    <Button htmlType="submit" type="primary">
                      Submit
                    </Button>
                  </Flex>
                </Form>
              </Space>
            </Card>
          );
        }
      })()}
    </SiderLayoutContent>
  );
}
