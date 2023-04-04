import { Row, Typography } from "antd";

import { ReactComponent as IconCheckMark } from "src/assets/icons/check.svg";
import { ReactComponent as IconCopy } from "src/assets/icons/copy-01.svg";

export function Detail({
  copyable,
  label,
  text,
}: {
  copyable?: { enabled: true; text?: string } | { enabled: false };
  label: string;
  text: string;
}) {
  return (
    <Row justify="space-between">
      <Typography.Text type="secondary">{label}</Typography.Text>

      <Typography.Text
        copyable={
          copyable?.enabled && {
            icon: [<IconCopy key={0} />, <IconCheckMark key={1} />],
            text: copyable.text || text,
          }
        }
        ellipsis
        style={{ textAlign: "right", width: 350 }}
      >
        {text}
      </Typography.Text>
    </Row>
  );
}
