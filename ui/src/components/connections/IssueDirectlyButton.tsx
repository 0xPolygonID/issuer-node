import { Button } from "antd";

import IconCreditCardPlus from "src/assets/icons/credit-card-plus.svg?react";

export function IssueDirectlyButton({ onClick }: { onClick: () => void }) {
  return (
    <Button icon={<IconCreditCardPlus />} onClick={onClick} type="primary">
      Issue directly
    </Button>
  );
}
