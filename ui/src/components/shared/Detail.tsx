import { Row, Typography } from "antd";

import { ReactComponent as IconCheckMark } from "src/assets/icons/check.svg";
import { ReactComponent as IconCopy } from "src/assets/icons/copy-01.svg";

export function Detail({
  copyable,
  ellipsisPosition,
  label,
  text,
}: {
  copyable?: boolean;
  ellipsisPosition?: number;
  label: string;
  text: string;
}) {
  return (
    <Row justify="space-between">
      <Typography.Text type="secondary">{label}</Typography.Text>

      <Typography.Text
        copyable={
          copyable && {
            icon: [<IconCopy key={0} />, <IconCheckMark key={1} />],
            text,
          }
        }
        ellipsis={ellipsisPosition ? { suffix: text.slice(-ellipsisPosition) } : true}
        style={{ textAlign: "right", width: 350 }}
      >
        {ellipsisPosition ? text.slice(0, text.length - ellipsisPosition) : text}
      </Typography.Text>
    </Row>
  );
}
