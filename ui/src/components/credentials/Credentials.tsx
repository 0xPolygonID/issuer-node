import { Button, Space } from "antd";
import { generatePath, useNavigate } from "react-router-dom";

import { ReactComponent as IconCreditCardPlus } from "src/assets/icons/credit-card-plus.svg";
import { CredentialsTable } from "src/components/credentials/CredentialsTable";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { ROUTES } from "src/routes";
import { CREDENTIALS, ISSUE_CREDENTIAL } from "src/utils/constants";

export function Credentials() {
  const navigate = useNavigate();

  return (
    <SiderLayoutContent
      description="Credentials that have been issued either directly or as credential links."
      extra={
        <Button
          icon={<IconCreditCardPlus />}
          onClick={() => navigate(generatePath(ROUTES.issueCredential.path))}
          type="primary"
        >
          {ISSUE_CREDENTIAL}
        </Button>
      }
      title={CREDENTIALS}
    >
      <Space direction="vertical">
        <CredentialsTable />
      </Space>
    </SiderLayoutContent>
  );
}
