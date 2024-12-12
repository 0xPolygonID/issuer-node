import { App, Button, Card, Flex, Form, Input, Space } from "antd";
import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";

import { useIdentityContext } from "../../contexts/Identity";
import { UpdateKey, getKey, updateKeyName } from "src/adapters/api/keys";
import CheckIcon from "src/assets/icons/check.svg?react";
import EditIcon from "src/assets/icons/edit-02.svg?react";
import CloseIcon from "src/assets/icons/x-close.svg?react";
import { Detail } from "src/components/shared/Detail";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { AppError, Key as KeyType } from "src/domain";
import { AsyncTask, hasAsyncTaskFailed, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { KEY_DETAILS, VALUE_REQUIRED } from "src/utils/constants";

export function Key() {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();

  const [nameEditable, setNameEditable] = useState(false);
  const { message } = App.useApp();
  const [form] = Form.useForm<UpdateKey>();

  const [key, setKey] = useState<AsyncTask<KeyType, AppError>>({
    status: "pending",
  });

  const { keyID } = useParams();

  const fetchKey = useCallback(
    async (signal?: AbortSignal) => {
      if (keyID) {
        setKey({ status: "loading" });

        const response = await getKey({
          env,
          identifier,
          keyID,
          signal,
        });

        if (response.success) {
          setKey({ data: response.data, status: "successful" });
        } else {
          if (!isAbortedError(response.error)) {
            setKey({ error: response.error, status: "failed" });
          }
        }
      }
    },
    [env, keyID, identifier]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchKey);

    return aborter;
  }, [fetchKey]);

  if (!keyID) {
    return <ErrorResult error="No key provided." />;
  }

  const handleEditName = (formValues: UpdateKey) => {
    return void updateKeyName({
      env,
      identifier,
      keyID,
      payload: { name: formValues.name.trim() },
    }).then((response) => {
      if (response.success) {
        void fetchKey().then(() => {
          setNameEditable(false);
          void message.success("Key edited successfully");
        });
      } else {
        void message.error(response.error.message);
      }
    });
  };

  return (
    <SiderLayoutContent
      description="View key details and edit name"
      showBackButton
      showDivider
      title={KEY_DETAILS}
    >
      {(() => {
        if (hasAsyncTaskFailed(key)) {
          return (
            <Card className="centered">
              <ErrorResult
                error={[
                  "An error occurred while downloading an key from the API:",
                  key.error.message,
                ].join("\n")}
              />
            </Card>
          );
        } else if (isAsyncTaskStarting(key)) {
          return (
            <Card className="centered">
              <LoadingResult />
            </Card>
          );
        } else {
          return (
            <Card
              className="centered"
              styles={{ header: { border: "none" } }}
              title={
                <Flex align="center" gap={8} style={{ paddingTop: "24px" }}>
                  {nameEditable ? (
                    <Form
                      form={form}
                      initialValues={{ name: key.data.name }}
                      onFinish={handleEditName}
                      style={{ width: "100%" }}
                    >
                      <Flex gap={16}>
                        <Form.Item
                          name="name"
                          rules={[{ message: VALUE_REQUIRED, required: true }]}
                          style={{ marginBottom: 0, width: "50%" }}
                        >
                          <Input placeholder="Enter name" />
                        </Form.Item>
                        <Flex gap={8}>
                          <Button icon={<CloseIcon />} onClick={() => setNameEditable(false)} />
                          <Button htmlType="submit" icon={<CheckIcon />} onClick={() => {}} />
                        </Flex>
                      </Flex>
                    </Form>
                  ) : (
                    <>
                      {key.data.name}
                      <Button
                        icon={<EditIcon />}
                        onClick={() => setNameEditable(true)}
                        size="small"
                        type="text"
                      />
                    </>
                  )}
                </Flex>
              }
            >
              <Card className="background-grey">
                <Space direction="vertical">
                  <Detail copyable label="Name" text={key.data.name} />

                  <Detail
                    copyable
                    copyableText={key.data.publicKey}
                    ellipsisPosition={5}
                    label="Public key"
                    text={key.data.publicKey}
                  />
                  <Detail copyable label="Type" text={key.data.keyType} />
                  <Detail label="Auth core clam" text={`${key.data.isAuthCredential}`} />
                </Space>
              </Card>
            </Card>
          );
        }
      })()}
    </SiderLayoutContent>
  );
}
