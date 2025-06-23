import { Button, Space } from "antd";
import { generatePath, useNavigate } from "react-router-dom";

import IconPlus from "src/assets/icons/plus.svg?react";
import { PaymentOptionsTable } from "src/components/payments/PaymentOptionsTable";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { ROUTES } from "src/routes";
import { PAYMENT_OPTIONS, PAYMENT_OPTIONS_ADD } from "src/utils/constants";

export function PaymentOptions() {
  const navigate = useNavigate();

  return (
    <SiderLayoutContent
      description="Description..."
      extra={
        <Button
          icon={<IconPlus />}
          onClick={() => navigate(generatePath(ROUTES.createPaymentOption.path))}
          type="primary"
        >
          {PAYMENT_OPTIONS_ADD}
        </Button>
      }
      title={PAYMENT_OPTIONS}
    >
      <Space direction="vertical" size="large">
        <PaymentOptionsTable />
      </Space>
    </SiderLayoutContent>
  );
}
