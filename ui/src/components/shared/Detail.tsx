import { Row, Tag, Typography } from "antd";

import { ReactComponent as IconCheckMark } from "src/assets/icons/check.svg";
import { ReactComponent as IconCopy } from "src/assets/icons/copy-01.svg";

export function Detail({
  copyable,
  ellipsisPosition,
  flavor = {
    type: "plain",
  },
  label,
  text,
}: {
  copyable?: boolean;
  ellipsisPosition?: number;
  flavor?:
    | {
        type: "plain";
      }
    | {
        color?: string;
        type: "tag";
      };
  label: string;
  text: string;
}) {
  const value = ellipsisPosition ? text.slice(0, text.length - ellipsisPosition) : text;
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
        {flavor.type === "plain" ? value : <Tag color={flavor.color}>{value}</Tag>}
      </Typography.Text>
    </Row>
  );
}
