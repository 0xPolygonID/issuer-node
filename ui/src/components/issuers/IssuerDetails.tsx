import { Button, Card, Flex, Form, Input, Space, message } from "antd";
import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { getIssuerDetails, updateIssuerDisplayName } from "src/adapters/api/issuers";
import { IssuerDetailsFormData } from "src/adapters/parsers/view";
import CheckIcon from "src/assets/icons/check.svg?react";
import EditIcon from "src/assets/icons/edit-02.svg?react";
import CloseIcon from "src/assets/icons/x-close.svg?react";
import { Detail } from "src/components/shared/Detail";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { AppError, IssuerInfo } from "src/domain";
import { AsyncTask, hasAsyncTaskFailed, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { ISSUER_DETAILS, VALUE_REQUIRED } from "src/utils/constants";
import { formatIdentifier } from "src/utils/forms";

export function IssuerDetails() {
  const env = useEnvContext();
  const [issuer, setIssuer] = useState<AsyncTask<IssuerInfo, AppError>>({
    status: "pending",
  });

  const [displayNameEditable, setDisplayNameEditable] = useState(false);
  const [messageAPI, messageContext] = message.useMessage();
  const [form] = Form.useForm<IssuerDetailsFormData>();

  const { issuerID: identifier } = useParams();

  const fetchIssuer = useCallback(
    async (signal?: AbortSignal) => {
      if (identifier) {
        setIssuer({ status: "loading" });

        const response = await getIssuerDetails({
          env,
          identifier,
          signal,
        });

        if (response.success) {
          setIssuer({ data: response.data, status: "successful" });
        } else {
          if (!isAbortedError(response.error)) {
            setIssuer({ error: response.error, status: "failed" });
          }
        }
      }
    },
    [env, identifier]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchIssuer);

    return aborter;
  }, [fetchIssuer]);

  if (!identifier) {
    return <ErrorResult error="No identifier provided." />;
  }

  const handleEditDisplayName = (formValues: IssuerDetailsFormData) =>
    void updateIssuerDisplayName({ displayName: formValues.displayName, env, identifier }).then(
      (response) => {
        if (response.success) {
          void fetchIssuer().then(() => {
            setDisplayNameEditable(false);
            void messageAPI.success("Identity edited successfully");
          });
        } else {
          void messageAPI.error(response.error.message);
        }
      }
    );

  return (
    <>
      {messageContext}
      <SiderLayoutContent
        description="View identity details and edit name"
        showBackButton
        showDivider
        title={ISSUER_DETAILS}
      >
        {(() => {
          if (hasAsyncTaskFailed(issuer)) {
            return (
              <Card className="centered">
                <ErrorResult
                  error={[
                    "An error occurred while downloading an issuer from the API:",
                    issuer.error.message,
                  ].join("\n")}
                />
              </Card>
            );
          } else if (isAsyncTaskStarting(issuer)) {
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
                styles={{
                  header: {
                    border: "none",
                  },
                }}
                title={
                  <Flex align="center" gap={8} style={{ paddingTop: "24px" }}>
                    {displayNameEditable ? (
                      <Form
                        form={form}
                        initialValues={{ displayName: issuer.data.displayName }}
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
                        {issuer.data.displayName}
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

                    <Detail label="Type" text={issuer.data.keyType} />
                    <Detail label="Credential status" text={issuer.data.credentialStatusType} />
                  </Space>
                </Card>
              </Card>
            );
          }
        })()}
      </SiderLayoutContent>
    </>
  );
}
