import { Modal, Typography, message } from "antd";

import { deleteLink } from "src/adapters/api/credentials";
import { ReactComponent as IconClose } from "src/assets/icons/x.svg";
import { useEnvContext } from "src/contexts/env";
import { CLOSE } from "src/utils/constants";

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

  const handleDeleteLink = () => {
    void deleteLink({ env, id }).then((response) => {
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
      cancelText={CLOSE}
      centered
      closable
      closeIcon={<IconClose />}
      maskClosable
      okButtonProps={{ danger: true }}
      okText="Delete"
      onCancel={onClose}
      onOk={handleDeleteLink}
      open
      title="Are you sure you want to delete this credential link?"
    >
      <Typography.Text type="secondary">
        Users will not be able to receive this credential any longer. This action cannot be undone.
      </Typography.Text>
    </Modal>
  );
}
