import { Flex, Modal, Typography } from "antd";
import { useState } from "react";
import IconTrash from "src/assets/icons/trash-01.svg?react";
import IconClose from "src/assets/icons/x.svg?react";
import { CLOSE, DELETE } from "src/utils/constants";

export function DeleteItem({ onOk, title }: { onOk: () => void; title: string }) {
  const [isModalOpen, setIsModalOpen] = useState(false);

  return (
    <>
      <Flex gap={8} onClick={() => setIsModalOpen(true)}>
        <IconTrash />
        <Typography.Text style={{ color: "inherit", fontSize: "inherit" }}>
          {DELETE}
        </Typography.Text>
      </Flex>
      <Modal
        cancelText={CLOSE}
        centered
        closable
        closeIcon={<IconClose />}
        maskClosable
        okButtonProps={{ danger: true }}
        okText={DELETE}
        onCancel={() => setIsModalOpen(false)}
        onOk={onOk}
        open={isModalOpen}
        title={title}
      >
        <Typography.Text type="secondary">This action cannot be undone.</Typography.Text>
      </Modal>
    </>
  );
}
