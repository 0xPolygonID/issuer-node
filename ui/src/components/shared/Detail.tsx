import { Row, Tag, Typography } from "antd";

import { z } from "zod";
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
  const isUrl = z.string().url().safeParse(text).success;
  const element = (
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
      {flavor.type === "plain" ? (
        value
      ) : (
        <Tag color={flavor.color} style={{ marginInlineEnd: "initial" }}>
          {value}
        </Tag>
      )}
    </Typography.Text>
  );

  return (
    <Row justify="space-between">
      <Typography.Text type="secondary">{label}</Typography.Text>
      {isUrl ? (
        <Typography.Link href={text} target="_blank">
          {element}
        </Typography.Link>
      ) : (
        element
      )}
    </Row>
  );
}
