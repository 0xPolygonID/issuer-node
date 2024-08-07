import { Modal, Space, Typography, message } from "antd";
import { useState } from "react";
import { useSearchParams } from "react-router-dom";

import { revokeCredential } from "src/adapters/api/credentials";
import IconClose from "src/assets/icons/x.svg?react";
import { useEnvContext } from "src/contexts/Env";
import { useIssuerContext } from "src/contexts/Issuer";
import { useIssuerStateContext } from "src/contexts/IssuerState";
import { Credential } from "src/domain";
import { CLOSE, REVOKE, REVOKED_SEARCH_PARAM } from "src/utils/constants";

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
  const { identifier } = useIssuerContext();
  const { notifyChange } = useIssuerStateContext();

  const [messageAPI, messageContext] = message.useMessage();

  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [, setSearchParams] = useSearchParams();

  const { revNonce: nonce } = credential;

  const handleRevokeCredential = () => {
    setIsLoading(true);

    void revokeCredential({ env, identifier, nonce }).then((response) => {
      if (response.success) {
        onClose();
        onRevoke();
        setSearchParams((previousParams) => {
          const params = new URLSearchParams(previousParams);
          params.set(REVOKED_SEARCH_PARAM, "true");
          return params;
        });
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
    </>
  );
}
