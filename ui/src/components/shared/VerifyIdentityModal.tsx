import { Modal, message } from "antd";
import { useState } from "react";

import { verifyIdentityRequest } from "src/adapters/api/requests";
import { ReactComponent as IconClose } from "src/assets/icons/x.svg";
import { useEnvContext } from "src/contexts/Env";
import { Request } from "src/domain";
import { CLOSE, VERIFY_IDENTITY } from "src/utils/constants";

export function VerifyIdentityModal({
  onClose,
  onVerify,
  request,
}: {
  onClose: () => void;
  onVerify: () => void;
  request: Request;
}) {
  const env = useEnvContext();

  const [messageAPI, messageContext] = message.useMessage();

  const [isLoading, setIsLoading] = useState<boolean>(false);

  const { id } = request;

  const handleVerifyIdentityRequest = () => {
    setIsLoading(true);

    void verifyIdentityRequest({ env, id }).then((response) => {
      if (response.success) {
        onClose();
        onVerify();
        void messageAPI.success(response.data.msg);
      } else {
        void messageAPI.error(response.error.message);
      }

      setIsLoading(false);
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
        okButtonProps={{ danger: true, loading: isLoading }}
        okText={VERIFY_IDENTITY}
        onCancel={onClose}
        onOk={handleVerifyIdentityRequest}
        open
        title="Are you sure you want to verify this credential?"
      >
        {/* <Space direction="vertical">
          <Typography.Text type="secondary">Are you Sure?</Typography.Text>
        </Space> */}
      </Modal>
    </>
  );
}
