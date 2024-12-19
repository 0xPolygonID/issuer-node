import { Avatar, Button, Result, Typography } from "antd";

import IconError from "src/assets/icons/alert-circle.svg?react";
import IconRetry from "src/assets/icons/refresh-ccw-01.svg?react";
import { DISPLAY_METHOD_DOCS_URL } from "src/utils/constants";

export function DisplayMethodErrorResult({
  labelRetry,
  message,
  onRetry,
}: {
  labelRetry: string;
  message: string;
  onRetry: () => void;
}) {
  return (
    <Result
      extra={[
        <Button icon={<IconRetry />} key={0} onClick={onRetry} type="primary">
          {labelRetry}
        </Button>,
      ]}
      icon={<Avatar className="avatar-color-error" icon={<IconError />} size={64} />}
      subTitle={message}
      title={
        <Typography.Paragraph style={{ fontSize: 16, textAlign: "center" }}>
          The display method is invalid. <br /> Please ensure it complies with the{" "}
          <Typography.Link href={DISPLAY_METHOD_DOCS_URL} style={{ fontSize: 16 }} target="_blank">
            documentation guidelines
          </Typography.Link>
        </Typography.Paragraph>
      }
    />
  );
}
