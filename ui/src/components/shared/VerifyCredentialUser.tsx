import { Modal, message } from "antd";
import { useState } from "react";

import { ReactComponent as IconClose } from "src/assets/icons/x.svg";
import { Credential } from "src/domain";
import { CLOSE, ISSUE_CREDENTIAL } from "src/utils/constants";

export function VerifyCredentialUser({
  credential,
  onClose,
}: {
  credential: Credential;
  onClose: () => void;
}) {
  const [messageAPI, messageContext] = message.useMessage();

  const [isLoading, setIsLoading] = useState<boolean>(false);

  const handleVerificationRequest = () => {
    console.log("-------------", messageAPI, credential);
    setIsLoading(true);

    // const payload = {
    //   Age: values.age,
    //   ProofId: values.adhaarID,
    //   ProofType: "Adhar",
    //   RequestType: "VerifyVC",
    //   RoleType: "Individual",
    //   schemaID: schema?.id,
    //   Source: "Manual",
    //   userDID: userDID,
    // };
    // await requestVC({
    //   env,
    //   payload,
    // }).then(void navigate(generatePath(ROUTES.request.path)));
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
        onOk={handleVerificationRequest}
        open
        title="Are you sure you want to send verification request?"
      >
        {/* <Space direction="vertical">
          <Typography.Text type="secondary">Are you Sure?</Typography.Text>
        </Space> */}
      </Modal>
    </>
  );
}
