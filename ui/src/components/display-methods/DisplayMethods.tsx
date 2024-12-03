import { Button, Space } from "antd";
import { generatePath, useNavigate } from "react-router-dom";

import IconPlus from "src/assets/icons/plus.svg?react";
import { DisplayMethodsTable } from "src/components/display-methods/DisplayMethodsTable";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { ROUTES } from "src/routes";
import { DISPLAY_METHODS, DISPLAY_METHOD_ADD } from "src/utils/constants";

export function DisplayMethods() {
  const navigate = useNavigate();

  return (
    <SiderLayoutContent
      description="Add new display methods and view details of existing display methods"
      extra={
        <Button
          icon={<IconPlus />}
          onClick={() => navigate(generatePath(ROUTES.createDisplayMethod.path))}
          type="primary"
        >
          {DISPLAY_METHOD_ADD}
        </Button>
      }
      title={DISPLAY_METHODS}
    >
      <Space direction="vertical" size="large">
        <DisplayMethodsTable />
      </Space>
    </SiderLayoutContent>
  );
}
