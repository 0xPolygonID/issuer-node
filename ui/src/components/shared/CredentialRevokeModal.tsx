import { Modal, Space, Typography, message } from "antd";
import { useState } from "react";

import { revokeCredential } from "src/adapters/api/credentials";
import { ReactComponent as IconClose } from "src/assets/icons/x.svg";
import { useEnvContext } from "src/contexts/Env";
import { useIssuerStateContext } from "src/contexts/IssuerState";
import { Credential } from "src/domain";
import { CLOSE, REVOKE } from "src/utils/constants";

export function CredentialRevokeModal({
  credential,
  onClose,
  onRevoke,
}: {
  credential: Credential;
  onClose: () => void;
  onRevoke: () => void;
}) {
  const env = useEnvContext();
  const { notifyChange } = useIssuerStateContext();
  const [isLoading, setIsLoading] = useState<boolean>(false);

  const { revNonce: nonce } = credential;

  const handleRevokeCredential = () => {
    setIsLoading(true);

    void revokeCredential({ env, nonce }).then((response) => {
      if (response.isSuccessful) {
        onClose();
        onRevoke();

        void notifyChange("revoke");
        void message.success(response.data);
      } else {
        setIsLoading(false);

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
      okButtonProps={{ danger: true, loading: isLoading }}
      okText={REVOKE}
      onCancel={onClose}
      onOk={handleRevokeCredential}
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
  );
}
