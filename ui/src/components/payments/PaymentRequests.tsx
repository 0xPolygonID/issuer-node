import { Space } from "antd";
import { PaymentRequestsTable } from "src/components/payments/PaymentRequestsTable";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { PAYMENT_REQUESTS } from "src/utils/constants";

export function PaymentRequests() {
  return (
    <SiderLayoutContent description="Description..." title={PAYMENT_REQUESTS}>
      <Space direction="vertical" size="large">
        <PaymentRequestsTable />
      </Space>
    </SiderLayoutContent>
  );
}
