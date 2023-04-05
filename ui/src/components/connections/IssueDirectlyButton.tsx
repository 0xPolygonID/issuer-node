import { Button } from "antd";

import { ReactComponent as IconCreditCardPlus } from "src/assets/icons/credit-card-plus.svg";

export function IssueDirectlyButton() {
  return (
    <Button icon={<IconCreditCardPlus />} type="primary">
      Issue directly
    </Button>
  );
}
