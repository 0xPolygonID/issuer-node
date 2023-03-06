import { Avatar, Col, Space, Typography } from "antd";

import { ReactComponent as IconSearch } from "src/assets/icons/search-lg.svg";

export function NoResults({ searchQuery }: { searchQuery: string | null }) {
  return (
    <Space direction="vertical" size="middle" style={{ padding: 24 }}>
      <Avatar className="avatar-color-cyan" icon={<IconSearch />} size={48} />

      <Typography.Text strong>No results found</Typography.Text>

      <Col push={6} span={12}>
        <Typography.Text>
          Your search <Typography.Text strong>{searchQuery && `"${searchQuery}" `}</Typography.Text>
          did not match any existing entries.
        </Typography.Text>

        <Typography.Paragraph> Please try searching for a different keyword.</Typography.Paragraph>
      </Col>
    </Space>
  );
}
