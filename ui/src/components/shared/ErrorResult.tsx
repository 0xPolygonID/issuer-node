import { Avatar, Button, Result } from "antd";

import IconError from "src/assets/icons/alert-circle.svg?react";
import IconRetry from "src/assets/icons/refresh-ccw-01.svg?react";
import { ERROR_MESSAGE } from "src/utils/constants";

export function ErrorResult({
  error,
  labelRetry = "Try again",
  onRetry = () => window.location.reload(),
}: {
  error: string;
  labelRetry?: string;
  onRetry?: () => void;
}) {
  return (
    <Result
      extra={[
        <Button icon={<IconRetry />} key={0} onClick={onRetry} type="primary">
          {labelRetry}
        </Button>,
      ]}
      icon={<Avatar className="avatar-color-error" icon={<IconError />} size={64} />}
      subTitle={error}
      title={ERROR_MESSAGE}
    />
  );
}
