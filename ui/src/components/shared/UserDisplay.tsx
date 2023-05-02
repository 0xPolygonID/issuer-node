import { Avatar, Row, Space, Typography } from "antd";
import { useEnvContext } from "src/contexts/Env";

export function UserDisplay() {
  const env = useEnvContext();

  return (
    <Space>
      <Avatar shape="square" size="large" src={env.issuer.logo} />

      <Row>
        <Typography.Text className="font-small" ellipsis strong>
          {env.issuer.name}
        </Typography.Text>
      </Row>
    </Space>
  );
}
