import { App, Flex, Space } from "antd";
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
  const { message } = App.useApp();

  const onDownload = () => {
    const canvas = ref.current?.querySelector("canvas");
    if (canvas) {
      downloadQRCanvas(canvas, fileName);
      void message.success("QR code downloaded successfully.");
    }
  };

  return (
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
  );
}
