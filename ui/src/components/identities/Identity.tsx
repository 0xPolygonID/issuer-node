import { App, Button, Card, Flex, Form, Input, Space } from "antd";
import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { useIdentityContext } from "../../contexts/Identity";
import { getIdentity, updateIdentityDisplayName } from "src/adapters/api/identities";
import { IdentityDetailsFormData } from "src/adapters/parsers/view";
import CheckIcon from "src/assets/icons/check.svg?react";
import EditIcon from "src/assets/icons/edit-02.svg?react";
import CloseIcon from "src/assets/icons/x-close.svg?react";
import { Detail } from "src/components/shared/Detail";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { AppError, IdentityDetails } from "src/domain";
import {
  AsyncTask,
  hasAsyncTaskFailed,
  isAsyncTaskDataAvailable,
  isAsyncTaskStarting,
} from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { IDENTITY_DETAILS, VALUE_REQUIRED } from "src/utils/constants";
import { formatIdentifier } from "src/utils/forms";

export function Identity() {
  const env = useEnvContext();
  const [identity, setIdentity] = useState<AsyncTask<IdentityDetails, AppError>>({
    status: "pending",
  });
  const { fetchIdentities, identityList } = useIdentityContext();

  const [displayNameEditable, setDisplayNameEditable] = useState(false);
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

  const handleEditDisplayName = (formValues: IdentityDetailsFormData) => {
    const isUnique =
      isAsyncTaskDataAvailable(identityList) &&
      !identityList.data.some(
        (identity) =>
          identity.identifier !== identifier && identity.displayName === formValues.displayName
      );

    if (!isUnique) {
      return void message.error(`${formValues.displayName} already exists`);
    }

    return void updateIdentityDisplayName({
      displayName: formValues.displayName.trim(),
      env,
      identifier,
    }).then((response) => {
      if (response.success) {
        void fetchIdentity().then(() => {
          setDisplayNameEditable(false);
          makeRequestAbortable(fetchIdentities);
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
            <Card
              className="centered"
              styles={{ header: { border: "none" } }}
              title={
                <Flex align="center" gap={8} style={{ paddingTop: "24px" }}>
                  {displayNameEditable ? (
                    <Form
                      form={form}
                      initialValues={{ displayName: identity.data.displayName }}
                      onFinish={handleEditDisplayName}
                      style={{ width: "100%" }}
                    >
                      <Flex gap={16}>
                        <Form.Item
                          name="displayName"
                          rules={[{ message: VALUE_REQUIRED, required: true }]}
                          style={{ marginBottom: 0, width: "50%" }}
                        >
                          <Input placeholder="Enter name" />
                        </Form.Item>
                        <Flex gap={8}>
                          <Button
                            icon={<CloseIcon />}
                            onClick={() => setDisplayNameEditable(false)}
                          />
                          <Button htmlType="submit" icon={<CheckIcon />} onClick={() => {}} />
                        </Flex>
                      </Flex>
                    </Form>
                  ) : (
                    <>
                      {identity.data.displayName}
                      <Button
                        icon={<EditIcon />}
                        onClick={() => setDisplayNameEditable(true)}
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
                  <Detail label="Credential status" text={identity.data.credentialStatusType} />
                </Space>
              </Card>
            </Card>
          );
        }
      })()}
    </SiderLayoutContent>
  );
}
