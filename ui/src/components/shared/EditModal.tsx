import { Divider, Modal } from "antd";
import { ReactNode } from "react";
import IconClose from "src/assets/icons/x.svg?react";

const MODAL_SIZES = {
  large: 800,
  small: 520,
};

export function EditModal({
  children,
  onClose,
  open,
  size = "small",
  title,
}: {
  children: ReactNode;
  onClose: () => void;
  open: boolean;
  size?: "small" | "large";
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
      title={title}
      width={MODAL_SIZES[size]}
    >
      <Divider />
      {children}
    </Modal>
  );
}
