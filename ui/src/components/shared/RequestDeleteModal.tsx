import { Alert, Button, Modal, Row, Space, Typography, message } from "antd";
import { useState } from "react";

import { deleteRequest, revokeRequest } from "src/adapters/api/requests";
import { ReactComponent as IconAlert } from "src/assets/icons/alert-triangle.svg";
import { ReactComponent as IconClose } from "src/assets/icons/x.svg";
import { useEnvContext } from "src/contexts/Env";
import { useIssuerStateContext } from "src/contexts/IssuerState";
import { Request } from "src/domain";
import { CLOSE, DELETE } from "src/utils/constants";

export function RequestDeleteModal({
  onClose,
  onDelete,
  request,
}: {
  onClose: () => void;
  onDelete: () => void;
  request: Request;
}) {
  const env = useEnvContext();
  const { notifyChange } = useIssuerStateContext();

  const [messageAPI, messageContext] = message.useMessage();

  const [isLoading, setIsLoading] = useState<boolean>(false);

  const { id, revNonce: nonce, revoked } = request;

  const handleDeleteRequest = () => {
    setIsLoading(true);

    void deleteRequest({ env, id }).then((response) => {
      if (response.success) {
        onClose();
        onDelete();

        void messageAPI.success(response.data.message);
      } else {
        void messageAPI.error(response.error.message);
      }

      setIsLoading(false);
    });
  };

  const handleRevokeRequest = () => {
    setIsLoading(true);

    void revokeRequest({ env, nonce }).then((response) => {
      if (response.success) {
        handleDeleteRequest();

        void notifyChange("revoke");
      } else {
        setIsLoading(false);

        void messageAPI.error(response.error.message);
      }
    });
  };

  return (
    <>
      {messageContext}

      <Modal
        centered
        closable
        closeIcon={<IconClose />}
        footer={
          <Row gutter={[8, 8]} justify="end">
            <Button onClick={onClose}>{CLOSE}</Button>

            <Button danger loading={isLoading} onClick={handleDeleteRequest} type="primary">
              {DELETE}
            </Button>

            {!revoked && (
              <Button danger loading={isLoading} onClick={handleRevokeRequest} type="primary">
                Delete & Revoke
              </Button>
            )}
          </Row>
        }
        maskClosable
        onCancel={onClose}
        open
        title="Are you sure you want to delete this credential?"
      >
        <Space direction="vertical">
          <Typography.Text type="secondary">
            Request data will be deleted from the database.
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
    </>
  );
}
