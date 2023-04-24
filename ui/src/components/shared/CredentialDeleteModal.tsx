import { Alert, Button, Modal, Space, Typography, message } from "antd";
import { useState } from "react";

import { deleteCredential, revokeCredential } from "src/adapters/api/credentials";
import { ReactComponent as IconAlert } from "src/assets/icons/alert-triangle.svg";
import { ReactComponent as IconClose } from "src/assets/icons/x.svg";
import { useEnvContext } from "src/contexts/Env";
import { useIssuerStateContext } from "src/contexts/IssuerState";
import { Credential } from "src/domain";
import { CLOSE } from "src/utils/constants";

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
  const { notifyChange } = useIssuerStateContext();

  const [isLoading, setIsLoading] = useState<boolean>(false);

  const { id, revNonce: nonce, revoked } = credential;

  const handleDeleteCredential = () => {
    setIsLoading(true);

    void deleteCredential({ env, id }).then((response) => {
      if (response.isSuccessful) {
        onClose();
        onDelete();

        void message.success(response.data);
      } else {
        void message.error(response.error.message);
      }

      setIsLoading(false);
    });
  };

  const handleRevokeCredential = () => {
    setIsLoading(true);

    void revokeCredential({ env, nonce }).then((response) => {
      if (response.isSuccessful) {
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
      footer={[
        <Button key="close" onClick={onClose}>
          {CLOSE}
        </Button>,
        <Button
          danger
          key="delete"
          loading={isLoading}
          onClick={handleDeleteCredential}
          type="primary"
        >
          Delete
        </Button>,
        !revoked && (
          <Button
            danger
            key="deleteRevoke"
            loading={isLoading}
            onClick={handleRevokeCredential}
            type="primary"
          >
            Delete & Revoke
          </Button>
        ),
      ]}
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
