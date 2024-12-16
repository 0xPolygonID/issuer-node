import { Divider, Modal } from "antd";
import { ReactNode } from "react";
import IconClose from "src/assets/icons/x.svg?react";
import { CLOSE, SAVE } from "src/utils/constants";

export function EditModal({
  children,
  onClose,
  onSubmit,
  open,
  title,
}: {
  children: ReactNode;
  onClose: () => void;
  onSubmit: () => void;
  open: boolean;
  title: string;
}) {
  return (
    <Modal
      cancelText={CLOSE}
      centered
      closable
      closeIcon={<IconClose />}
      maskClosable
      okText={SAVE}
      onCancel={onClose}
      onOk={onSubmit}
      open={open}
      style={{ maxWidth: 600 }}
      title={title}
    >
      <Divider />
      {children}
      <Divider />
    </Modal>
  );
}
