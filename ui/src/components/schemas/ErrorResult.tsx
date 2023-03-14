import { Avatar, Button, Col, Space, Typography } from "antd";

import { ReactComponent as IconError } from "src/assets/icons/alert-circle.svg";
import { ReactComponent as IconRetry } from "src/assets/icons/refresh-ccw-01.svg";
import { ERROR_MESSAGE } from "src/utils/constants";

export function ErrorResult({ error }: { error: string }) {
  const onReload = () => window.location.reload();

  return (
    <Space align="center" direction="vertical" size="middle" style={{ padding: 24 }}>
      <Avatar className="avatar-color-error" icon={<IconError />} size={48} />

      <Typography.Text strong style={{ fontSize: 16 }}>
        {ERROR_MESSAGE}
      </Typography.Text>

      <Col>
        <Typography.Text style={{ whiteSpace: "pre-line" }}>{error}</Typography.Text>
      </Col>

      <Col>
        <Button icon={<IconRetry />} key={0} onClick={onReload} type="link">
          Reload
        </Button>
      </Col>
    </Space>
  );
}
