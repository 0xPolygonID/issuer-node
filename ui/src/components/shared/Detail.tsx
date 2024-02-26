import { Col, Grid, Row, Tag, TagProps, Typography } from "antd";

import IconCheckMark from "src/assets/icons/check.svg?react";
import IconCopy from "src/assets/icons/copy-01.svg?react";

export function Detail({
  copyable,
  ellipsisPosition,
  href,
  label,
  tag,
  text,
}: {
  copyable?: boolean;
  ellipsisPosition?: number;
  href?: string;
  label: string;
  tag?: TagProps;
  text: string;
}) {
  const { sm } = Grid.useBreakpoint();
  const value = ellipsisPosition ? text.slice(0, text.length - ellipsisPosition) : text;
  const element = (
    <Typography.Text
      copyable={
        copyable && {
          icon: [<IconCopy key={0} />, <IconCheckMark key={1} />],
          text,
        }
      }
      ellipsis={ellipsisPosition ? { suffix: text.slice(-ellipsisPosition) } : true}
      style={{ textAlign: sm ? "right" : "left", width: "100%" }}
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
      <Col sm={10} xs={24}>
        <Typography.Text type="secondary">{label}</Typography.Text>
      </Col>
      <Col sm={14} xs={24}>
        {href ? (
          <Typography.Link ellipsis href={href} target="_blank">
            {element}
          </Typography.Link>
        ) : (
          element
        )}
      </Col>
    </Row>
  );
}
