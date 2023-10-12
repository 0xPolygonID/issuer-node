import { Modal, message } from "antd";
import { useState } from "react";

import { getSchema, issueCredentialRequest } from "src/adapters/api/requests";
import { ReactComponent as IconClose } from "src/assets/icons/x.svg";
import { useEnvContext } from "src/contexts/Env";
import { Request } from "src/domain";
import { CLOSE, ISSUE_CREDENTIAL } from "src/utils/constants";

export function IssueCredentialUser({
  onClose,
  request,
}: {
  onClose: () => void;
  request: Request;
}) {
  const env = useEnvContext();

  const [messageAPI, messageContext] = message.useMessage();

  const [isLoading, setIsLoading] = useState<boolean>(false);

  const handleCredentialRequest = () => {
    setIsLoading(true);
    const schemaID = request.schemaID;

    void getSchema({ env, schemaID }).then((schemaReponse) => {
      if (schemaReponse.success) {
        const dataSchema = {
          credentialSchema: schemaReponse.data.url,
          credentialSubject: {
            "Adhar-number": parseInt(request.proof_id),
            Age: 1,
            id: request.userDID,
          },
          expiration: request.created_at,
          mtProof: false,
          requestId: request.id,
          signatureProof: true,
          type: schemaReponse.data.type,
        };
        void issueCredentialRequest({ dataSchema, env }).then((response) => {
          console.log("--------------response", response);

          if (response.success) {
            onClose();

            void messageAPI.success(response.data.message);
          } else {
            setIsLoading(false);

            void messageAPI.error(response.error.message);
          }
        });
      } else {
        setIsLoading(false);
        void messageAPI.error(schemaReponse.error.message);
      }
    });
    return;
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
        okText={ISSUE_CREDENTIAL}
        onCancel={onClose}
        onOk={handleCredentialRequest}
        open
        title="Are you sure you want to issue this credential?"
      >
        {/* <Space direction="vertical">
          <Typography.Text type="secondary">Are you Sure?</Typography.Text>
        </Space> */}
      </Modal>
    </>
  );
}
