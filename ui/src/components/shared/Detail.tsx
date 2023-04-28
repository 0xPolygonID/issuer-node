import { Row, Tag, TagProps, Typography } from "antd";

import { ReactComponent as IconCheckMark } from "src/assets/icons/check.svg";
import { ReactComponent as IconCopy } from "src/assets/icons/copy-01.svg";

export function Detail({
  copyable,
  ellipsisPosition,
  label,
  tag,
  text,
}: {
  copyable?: boolean;
  ellipsisPosition?: number;
  label: string;
  tag?: TagProps;
  text: string;
}) {
  const value = ellipsisPosition ? text.slice(0, text.length - ellipsisPosition) : text;
  const isUrl = text.startsWith("http://") || text.startsWith("https://");
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
      {tag ? (
        <Tag {...tag} style={{ marginInlineEnd: "initial" }}>
          {value}
        </Tag>
      ) : (
        value
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
