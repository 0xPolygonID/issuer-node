import { Button } from "antd";

import { ReactComponent as IconCreditCardPlus } from "src/assets/icons/credit-card-plus.svg";

export function IssueDirectlyButton({ onClick }: { onClick: () => void }) {
  return (
    <Button icon={<IconCreditCardPlus />} onClick={onClick} type="primary">
      Issue directly
    </Button>
  );
}
