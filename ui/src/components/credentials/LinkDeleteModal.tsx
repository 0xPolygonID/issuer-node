import { Modal, Typography, message } from "antd";

import { deleteLink } from "src/adapters/api/credentials";
import IconClose from "src/assets/icons/x.svg?react";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { CLOSE, DELETE } from "src/utils/constants";

export function LinkDeleteModal({
  id,
  onClose,
  onDelete,
}: {
  id: string;
  onClose: () => void;
  onDelete: () => void;
}) {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();

  const [messageAPI, messageContext] = message.useMessage();

  const handleDeleteLink = () => {
    void deleteLink({ env, id, identifier }).then((response) => {
      if (response.success) {
        onClose();
        onDelete();

        void messageAPI.success(response.data.message);
      } else {
        void messageAPI.error(response.error.message);
      }
    });
  };

  return (
    <>
      {messageContext}

      <Modal
        cancelText={CLOSE}
        centered
        closable
        closeIcon={<IconClose />}
        maskClosable
        okButtonProps={{ danger: true }}
        okText={DELETE}
        onCancel={onClose}
        onOk={handleDeleteLink}
        open
        title="Are you sure you want to delete this credential link?"
      >
        <Typography.Text type="secondary">
          Users will not be able to receive this credential any longer. This action cannot be
          undone.
        </Typography.Text>
      </Modal>
    </>
  );
}
