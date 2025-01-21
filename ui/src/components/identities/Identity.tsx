import { App, Button, Card, Divider, Flex, Form, Input, Space, Typography } from "antd";
import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";

import { getIdentity, updateIdentityDisplayName } from "src/adapters/api/identities";
import { IdentityDetailsFormData } from "src/adapters/parsers/view";
import EditIcon from "src/assets/icons/edit-02.svg?react";
import { IdentityAuthCredentials } from "src/components/identities/IdentityAuthCredentials";
import { Detail } from "src/components/shared/Detail";
import { EditModal } from "src/components/shared/EditModal";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { AppError, IdentityDetails } from "src/domain";
import { AsyncTask, hasAsyncTaskFailed, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { IDENTITY_DETAILS, SAVE, VALUE_REQUIRED } from "src/utils/constants";
import { formatIdentifier } from "src/utils/forms";

export function Identity() {
  const env = useEnvContext();
  const { fetchIdentities } = useIdentityContext();
  const [identity, setIdentity] = useState<AsyncTask<IdentityDetails, AppError>>({
    status: "pending",
  });

  const [isEditModalOpen, setIsEditModalOpen] = useState(false);

  const { message } = App.useApp();
  const [form] = Form.useForm<IdentityDetailsFormData>();

  const { identityID: identifier } = useParams();

  const fetchIdentity = useCallback(
    async (signal?: AbortSignal) => {
      if (identifier) {
        setIdentity({ status: "loading" });

        const response = await getIdentity({
          env,
          identifier,
          signal,
        });

        if (response.success) {
          setIdentity({ data: response.data, status: "successful" });
        } else {
          if (!isAbortedError(response.error)) {
            setIdentity({ error: response.error, status: "failed" });
          }
        }
      }
    },
    [env, identifier]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchIdentity);

    return aborter;
  }, [fetchIdentity]);

  if (!identifier) {
    return <ErrorResult error="No identifier provided." />;
  }

  const handleEdit = (values: { displayName: string }) => {
    const { displayName } = values;
    void updateIdentityDisplayName({
      displayName,
      env,
      identifier,
    }).then((response) => {
      if (response.success) {
        void fetchIdentity().then(() => {
          setIsEditModalOpen(false);
          void fetchIdentities();
          void message.success("Identity edited successfully");
        });
      } else {
        void message.error(response.error.message);
      }
    });
  };

  return (
    <SiderLayoutContent
      description="View identity details and edit name"
      showBackButton
      showDivider
      title={IDENTITY_DETAILS}
    >
      {(() => {
        if (hasAsyncTaskFailed(identity)) {
          return (
            <Card className="centered">
              <ErrorResult
                error={[
                  "An error occurred while downloading an identity from the API:",
                  identity.error.message,
                ].join("\n")}
              />
            </Card>
          );
        } else if (isAsyncTaskStarting(identity)) {
          return (
            <Card className="centered">
              <LoadingResult />
            </Card>
          );
        } else {
          const [, method = "", blockchain = "", network = ""] = identifier.split(":");
          return (
            <>
              <Card
                className="centered"
                title={
                  <Flex align="center" gap={8} justify="space-between">
                    <Typography.Text style={{ fontWeight: 600 }}>
                      {identity.data.displayName}
                    </Typography.Text>
                    <Flex gap={8}>
                      <Button
                        icon={<EditIcon />}
                        onClick={() => setIsEditModalOpen(true)}
                        style={{ flexShrink: 0 }}
                        type="text"
                      />
                    </Flex>
                  </Flex>
                }
              >
                <Card className="background-grey">
                  <Space direction="vertical">
                    <Detail
                      copyable
                      copyableText={identifier}
                      label="Identifier"
                      text={formatIdentifier(identifier)}
                    />

                    <Detail label="Method" text={method} />

                    <Detail label="Blockchain" text={blockchain} />

                    <Detail label="Network" text={network} />

                    <Detail label="Type" text={identity.data.keyType} />
                  </Space>
                </Card>
              </Card>

              <Divider />

              {identity.data.authCredentialsIDs.length && (
                <IdentityAuthCredentials
                  identityID={identifier}
                  IDs={identity.data.authCredentialsIDs}
                />
              )}

              <EditModal
                onClose={() => setIsEditModalOpen(false)}
                open={isEditModalOpen}
                title="Edit identity"
              >
                <Form
                  form={form}
                  initialValues={{ displayName: identity.data.displayName }}
                  layout="vertical"
                  onFinish={handleEdit}
                >
                  <Form.Item
                    name="displayName"
                    rules={[{ message: VALUE_REQUIRED, required: true }]}
                  >
                    <Input placeholder="Enter name" />
                  </Form.Item>

                  <Divider />

                  <Flex justify="flex-end">
                    <Button htmlType="submit" type="primary">
                      {SAVE}
                    </Button>
                  </Flex>
                </Form>
              </EditModal>
            </>
          );
        }
      })()}
    </SiderLayoutContent>
  );
}
