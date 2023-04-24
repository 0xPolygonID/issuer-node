import { Button, Modal, Typography } from "antd";
import { QRCodeSVG } from "qrcode.react";

import { ReactComponent as IconClose } from "src/assets/icons/x.svg";
import { CLOSE } from "src/utils/constants";

export function ClaimCredentialModal({
  onClose,
  qrCode,
}: {
  onClose: () => void;
  qrCode?: unknown;
}) {
  return (
    <Modal
      centered
      closable
      closeIcon={<IconClose />}
      footer={[
        <Button key="close" onClick={onClose}>
          {CLOSE}
        </Button>,
      ]}
      maskClosable
      onCancel={onClose}
      open
      title="Missed the notification?"
    >
      <Typography.Text type="secondary">
        Scan the QR to add the credential to your wallet
      </Typography.Text>

      <QRCodeSVG
        className="full-width"
        includeMargin
        level="H"
        style={{ height: 300 }}
        value={JSON.stringify(qrCode)}
      />
    </Modal>
  );
}
