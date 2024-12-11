import { Button, Space } from "antd";
import { generatePath, useNavigate } from "react-router-dom";

import IconPlus from "src/assets/icons/plus.svg?react";
import { KeysTable } from "src/components/keys/KeysTable";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { ROUTES } from "src/routes";
import { KEYS, KEY_ADD } from "src/utils/constants";

export function Keys() {
  const navigate = useNavigate();

  return (
    <SiderLayoutContent
      description="Description..."
      extra={
        <Button
          icon={<IconPlus />}
          onClick={() => navigate(generatePath(ROUTES.createKey.path))}
          type="primary"
        >
          {KEY_ADD}
        </Button>
      }
      title={KEYS}
    >
      <Space direction="vertical" size="large">
        <KeysTable />
      </Space>
    </SiderLayoutContent>
  );
}
