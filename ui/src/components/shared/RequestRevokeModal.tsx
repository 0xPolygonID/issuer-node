import { Modal, Space, Typography, message } from "antd";
import { useState } from "react";

import { revokeCredential } from "src/adapters/api/credentials";
import { ReactComponent as IconClose } from "src/assets/icons/x.svg";
import { useEnvContext } from "src/contexts/Env";
import { useIssuerStateContext } from "src/contexts/IssuerState";
import { Request } from "src/domain";
import { CLOSE, REVOKE } from "src/utils/constants";

export function RequestRevokeModal({
  onClose,
  onRevoke,
  request,
}: {
  onClose: () => void;
  onRevoke: () => void;
  request: Request;
}) {
  const env = useEnvContext();
  const { notifyChange } = useIssuerStateContext();

  const [messageAPI, messageContext] = message.useMessage();

  const [isLoading, setIsLoading] = useState<boolean>(false);

  const { revNonce: nonce } = request;

  const handleRevokeRequest = () => {
    setIsLoading(true);

    void revokeCredential({ env, nonce }).then((response) => {
      if (response.success) {
        onClose();
        onRevoke();

        void notifyChange("revoke");
        void messageAPI.success(response.data.message);
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
        cancelText={CLOSE}
        centered
        closable
        closeIcon={<IconClose />}
        maskClosable
        okButtonProps={{ danger: true, loading: isLoading }}
        okText={REVOKE}
        onCancel={onClose}
        onOk={handleRevokeRequest}
        open
        title="Are you sure you want to revoke this credential?"
      >
        <Space direction="vertical">
          <Typography.Text type="secondary">
            Revoking of a credential must be accompanied by publishing of issuer state in order for
            the action to be effective. This action cannot be undone.
          </Typography.Text>
        </Space>
      </Modal>
    </>
  );
}
