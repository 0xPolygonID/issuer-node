import { Flex, Space, message } from "antd";
import { QRCodeCanvas } from "qrcode.react";
import { useRef } from "react";
import { downloadQRCanvas } from "src/utils/browser";

export function DownloadQRLink({
  button,
  fileName,
  hidden,
  link,
}: {
  button: JSX.Element;
  fileName: string;
  hidden: boolean;
  link: string;
}) {
  const ref = useRef<HTMLDivElement | null>(null);
  const [messageAPI, messageContext] = message.useMessage();

  const onDownload = () => {
    const canvas = ref.current?.querySelector("canvas");
    if (canvas) {
      downloadQRCanvas(canvas, fileName);
      void messageAPI.success("QR code downloaded successfully.");
    }
  };

  return (
    <>
      {messageContext}
      <Flex align="center" gap={12} ref={ref} vertical>
        <QRCodeCanvas
          className="qr-code full-width"
          itemRef="ref"
          level="H"
          size={300}
          style={{ ...(hidden && { display: "none" }), height: 300 }}
          value={link}
        />
        <Space onClick={onDownload}>{button}</Space>
      </Flex>
    </>
  );
}
