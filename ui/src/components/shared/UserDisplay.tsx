import { Avatar, Row, Space, Typography } from "antd";
import { useEnvContext } from "src/contexts/Env";

export function UserDisplay() {
  const { issuer } = useEnvContext();

  return (
    <Space>
      <Avatar shape="square" size="large" src={issuer.logo} />

      <Row>
        <Typography.Text className="font-small" ellipsis strong>
          {issuer.name}
        </Typography.Text>
      </Row>
    </Space>
  );
}
