import { Checkbox, Divider, Modal, Space, Typography, message } from "antd";
import { CheckboxChangeEvent } from "antd/es/checkbox";
import { useState } from "react";

import { deleteConnection } from "src/adapters/api/connections";
import { ReactComponent as IconClose } from "src/assets/icons/x.svg";
import { useEnvContext } from "src/contexts/env";

export function ConnectionDeleteModal({
  id,
  onClose,
  onDelete,
}: {
  id: string;
  onClose: () => void;
  onDelete: () => void;
}) {
  const env = useEnvContext();
  const [revokeCredentials, setRevokeCredentials] = useState<boolean>(false);
  const [deleteCredentials, setDeleteCredentials] = useState<boolean>(false);

  const handleDeleteConnection = () => {
    void deleteConnection({ deleteCredentials, env, id, revokeCredentials }).then((response) => {
      if (response.isSuccessful) {
        onClose();
        onDelete();

        void message.success(response.data);
      } else {
        void message.error(response.error.message);
      }
    });
  };

  return (
    <Modal
      cancelText="Close"
      centered
      closable
      closeIcon={<IconClose />}
      maskClosable
      okButtonProps={{ danger: true }}
      okText="Delete connection"
      onCancel={onClose}
      onOk={handleDeleteConnection}
      open
      title="Are you sure you want to delete this connection?"
    >
      <Typography.Text type="secondary">
        Identity will be deleted from your connections.
      </Typography.Text>
      <Divider />
      <Space direction="vertical">
        <Typography.Text strong>Would you also like to:</Typography.Text>
        <Checkbox
          onChange={({ target: { checked } }: CheckboxChangeEvent) => setRevokeCredentials(checked)}
        >
          <Typography.Text>Revoke all credentials for this connection.</Typography.Text>
          <Typography.Paragraph type="secondary">
            Revoking must be accompanied by publishing of issuer state in order for the action to be
            effective.
          </Typography.Paragraph>
        </Checkbox>
        <Checkbox
          onChange={({ target: { checked } }: CheckboxChangeEvent) => setDeleteCredentials(checked)}
        >
          <Typography.Text>Delete all credentials for this connection.</Typography.Text>
          <Typography.Paragraph type="secondary">
            Credential data will be deleted from the database.
          </Typography.Paragraph>
        </Checkbox>
      </Space>
    </Modal>
  );
}
