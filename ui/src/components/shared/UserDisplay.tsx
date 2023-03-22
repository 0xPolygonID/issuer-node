import { Avatar, Row, Space, Typography } from "antd";
import { env } from "src/adapters/parsers/env";
import { IMAGE_PLACEHOLDER_PATH } from "src/utils/constants";

export function UserDisplay() {
  return (
    <Space>
      <Avatar shape="square" size="large" src={env.issuer.logo || IMAGE_PLACEHOLDER_PATH} />

      <Row>
        <Typography.Text className="font-small" ellipsis strong>
          {env.issuer.name}
        </Typography.Text>
      </Row>
    </Space>
  );
}
