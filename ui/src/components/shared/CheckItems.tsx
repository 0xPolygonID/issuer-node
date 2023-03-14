import { Avatar, Space, Typography } from "antd";

import { ReactComponent as IconCheck } from "src/assets/icons/check.svg";

export function CheckItems({ items }: { items: string | string[] }) {
  const contents = typeof items === "string" ? [items] : items;

  return (
    <Space direction="vertical" size="middle">
      {contents.map((item, key) => (
        <Space className="check-items" key={key}>
          <Avatar className="avatar-color-cyan" icon={<IconCheck />} size={24} />

          <Typography.Text type="secondary">{item}</Typography.Text>
        </Space>
      ))}
    </Space>
  );
}
