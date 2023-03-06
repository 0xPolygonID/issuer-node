import { Avatar, Row, Space, Typography } from "antd";
import { IMAGE_PLACEHOLDER_PATH, ISSUER_LOGO, ISSUER_NAME } from "src/utils/constants";

export function UserDisplay() {
  return (
    <Space>
      <Avatar shape="square" size="large" src={ISSUER_LOGO || IMAGE_PLACEHOLDER_PATH} />

      <Row>
        <Typography.Text className="font-small" ellipsis strong>
          {ISSUER_NAME}
        </Typography.Text>
      </Row>
    </Space>
  );
}
