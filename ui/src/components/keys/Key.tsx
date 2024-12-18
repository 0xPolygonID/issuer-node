import { App, Button, Card, Flex, Form, Input, Space, Typography } from "antd";
import { useCallback, useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";

import { useIdentityContext } from "../../contexts/Identity";
import { UpdateKey, deleteKey, getKey, updateKeyName } from "src/adapters/api/keys";
import EditIcon from "src/assets/icons/edit-02.svg?react";
import IconTrash from "src/assets/icons/trash-01.svg?react";
import { Detail } from "src/components/shared/Detail";
import { EditModal } from "src/components/shared/EditModal";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { AppError, Key as KeyType } from "src/domain";
import { ROUTES } from "src/routes";
import { AsyncTask, hasAsyncTaskFailed, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { KEY_DETAILS, VALUE_REQUIRED } from "src/utils/constants";

export function Key() {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();
  const { message } = App.useApp();
  const navigate = useNavigate();
  const [form] = Form.useForm<UpdateKey>();

  const [isEditModalOpen, setIsEditModalOpen] = useState(false);

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

  const handleEdit = () => {
    const { name } = form.getFieldsValue();
    void updateKeyName({
      env,
      identifier,
      keyID,
      payload: { name: name.trim() },
    }).then((response) => {
      setIsEditModalOpen(false);
      if (response.success) {
        void fetchKey().then(() => {
          void message.success("Key edited successfully");
        });
      } else {
        void message.error(response.error.message);
      }
    });
  };

  const handleDeleteKey = () => {
    void deleteKey({ env, identifier, keyID }).then((response) => {
      if (response.success) {
        navigate(ROUTES.keys.path);
        void message.success(response.data.message);
      } else {
        void message.error(response.error.message);
      }
    });
  };

  return (
    <SiderLayoutContent
      description="View and edit key details"
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
            <>
              <Card
                className="centered"
                title={
                  <Flex align="center" gap={8} justify="space-between">
                    <Typography.Text style={{ fontWeight: 600 }}>{key.data.name}</Typography.Text>
                    <Flex gap={8}>
                      <Button
                        icon={<EditIcon />}
                        onClick={() => setIsEditModalOpen(true)}
                        style={{ flexShrink: 0 }}
                        type="text"
                      />
                      {!key.data.isAuthCredential && (
                        <Button
                          danger
                          icon={<IconTrash />}
                          onClick={handleDeleteKey}
                          style={{ flexShrink: 0 }}
                          type="text"
                        />
                      )}
                    </Flex>
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
                    <Detail label="Auth credential" text={`${key.data.isAuthCredential}`} />
                  </Space>
                </Card>
              </Card>

              <EditModal
                onClose={() => setIsEditModalOpen(false)}
                onSubmit={handleEdit}
                open={isEditModalOpen}
                title="Edit key"
              >
                <Form form={form} initialValues={{ name: key.data.name }} layout="vertical">
                  <Form.Item
                    label="Name"
                    name="name"
                    rules={[{ message: VALUE_REQUIRED, required: true }]}
                  >
                    <Input placeholder="Enter name" />
                  </Form.Item>
                </Form>
              </EditModal>
            </>
          );
        }
      })()}
    </SiderLayoutContent>
  );
}