import { Alert, App, Button, Flex, Modal, Space, Typography } from "antd";
import { useState } from "react";

import { deleteCredential, revokeCredential } from "src/adapters/api/credentials";
import IconAlert from "src/assets/icons/alert-triangle.svg?react";
import IconClose from "src/assets/icons/x.svg?react";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { useIssuerStateContext } from "src/contexts/IssuerState";
import { Credential } from "src/domain";
import { CLOSE, DELETE } from "src/utils/constants";

export function CredentialDeleteModal({
  credential,
  onClose,
  onDelete,
}: {
  credential: Credential;
  onClose: () => void;
  onDelete: () => void;
}) {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();
  const { notifyChange } = useIssuerStateContext();

  const { message } = App.useApp();

  const [isLoading, setIsLoading] = useState<boolean>(false);

  const { id, revNonce: nonce, revoked } = credential;

  const handleDeleteCredential = () => {
    setIsLoading(true);

    void deleteCredential({ env, id, identifier }).then((response) => {
      if (response.success) {
        onClose();
        onDelete();

        void message.success(response.data.message);
      } else {
        void message.error(response.error.message);
      }

      setIsLoading(false);
    });
  };

  const handleRevokeCredential = () => {
    setIsLoading(true);

    void revokeCredential({ env, identifier, nonce }).then((response) => {
      if (response.success) {
        handleDeleteCredential();

        void notifyChange("revoke");
      } else {
        setIsLoading(false);

        void message.error(response.error.message);
      }
    });
  };

  return (
    <Modal
      centered
      closable
      closeIcon={<IconClose />}
      footer={
        <Flex gap={8} justify="end">
          <Button onClick={onClose}>{CLOSE}</Button>

          <Button danger loading={isLoading} onClick={handleDeleteCredential} type="primary">
            {DELETE}
          </Button>

          {!revoked && (
            <Button danger loading={isLoading} onClick={handleRevokeCredential} type="primary">
              Delete & Revoke
            </Button>
          )}
        </Flex>
      }
      maskClosable
      onCancel={onClose}
      open
      title="Are you sure you want to delete this credential?"
    >
      <Space direction="vertical">
        <Typography.Text type="secondary">
          Credential data will be deleted from the database.
        </Typography.Text>

        {!revoked && (
          <Alert
            description={
              <Typography.Text type="warning">
                Revoking requires publishing the issuer state. This action cannot be undone.
              </Typography.Text>
            }
            icon={<IconAlert />}
            message={
              <Typography.Text strong type="warning">
                If a credential is deleted but not revoked, it can still be used by end users.
              </Typography.Text>
            }
            showIcon
            type="warning"
          />
        )}
      </Space>
    </Modal>
  );
}
