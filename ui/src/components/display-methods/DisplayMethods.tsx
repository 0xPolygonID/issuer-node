import { Button, Space, Typography } from "antd";
import { generatePath, useNavigate } from "react-router-dom";

import IconPlus from "src/assets/icons/plus.svg?react";
import { DisplayMethodsTable } from "src/components/display-methods/DisplayMethodsTable";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { ROUTES } from "src/routes";
import { DISPLAY_METHODS, DISPLAY_METHOD_ADD, DISPLAY_METHOD_DOCS_URL } from "src/utils/constants";

export function DisplayMethods() {
  const navigate = useNavigate();

  return (
    <SiderLayoutContent
      description={
        <Typography.Text>
          Add new display methods or view detailed information about existing ones. For more
          guidance, refer to our{" "}
          <Typography.Link href={DISPLAY_METHOD_DOCS_URL} target="_blank">
            documentation
          </Typography.Link>
          .
        </Typography.Text>
      }
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
