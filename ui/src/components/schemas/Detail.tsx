import { Row, Typography } from "antd";

import { ReactComponent as IconCheckMark } from "src/assets/icons/check.svg";
import { ReactComponent as IconCopy } from "src/assets/icons/copy-01.svg";

export function Detail({
  copyable = false,
  data,
  label,
}: {
  copyable?: boolean;
  data: string;
  label: string;
  maxWidth?: number;
}) {
  return (
    <Row justify="space-between">
      <Typography.Text type="secondary">{label}</Typography.Text>

      <Typography.Text
        copyable={copyable && { icon: [<IconCopy key={0} />, <IconCheckMark key={1} />] }}
        ellipsis
        style={{ textAlign: "right", width: 350 }}
      >
        {data}
      </Typography.Text>
    </Row>
  );
}
