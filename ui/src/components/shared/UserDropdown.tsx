import { Avatar, Row, Space, Typography } from "antd";
import { IMAGE_PLACEHOLDER_PATH, ISSUER_LOGO, ISSUER_NAME } from "src/utils/constants";

export function UserDropdown() {
  return (
    <Space>
      <Avatar shape="square" size="large" src={ISSUER_LOGO || IMAGE_PLACEHOLDER_PATH} />

      <Row className="dropdown-info">
        <Typography.Text className="font-small" ellipsis strong>
          {ISSUER_NAME}
        </Typography.Text>
      </Row>
    </Space>
  );
}
