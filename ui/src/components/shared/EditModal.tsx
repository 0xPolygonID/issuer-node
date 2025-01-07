import { Divider, Modal } from "antd";
import { ReactNode } from "react";
import IconClose from "src/assets/icons/x.svg?react";

export function EditModal({
  children,
  onClose,
  open,
  title,
}: {
  children: ReactNode;
  onClose: () => void;
  open: boolean;
  title: string;
}) {
  return (
    <Modal
      centered
      closable
      closeIcon={<IconClose />}
      destroyOnClose
      footer={null}
      maskClosable
      onCancel={onClose}
      open={open}
      style={{ maxWidth: 600 }}
      title={title}
    >
      <Divider />
      {children}
    </Modal>
  );
}
