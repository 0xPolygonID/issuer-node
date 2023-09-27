import { Button, Space } from "antd";
import { generatePath, useNavigate } from "react-router-dom";

import { RequestsTable } from "./RequestsTable";
import { ReactComponent as IconCreditCardPlus } from "src/assets/icons/credit-card-plus.svg";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { ROUTES } from "src/routes";
import {
  ISSUED_CREDENTIALS,
  REQUESTS,
  REQUEST_FOR_VC,
  REQUEST_FOR_VC_CREDS,
} from "src/utils/constants";

export function Request() {
  const navigate = useNavigate();
  const User = localStorage.getItem("user");

  return (
    <SiderLayoutContent
      description="Description for Request tab"
      extra={
        <Button
          icon={<IconCreditCardPlus />}
          onClick={() => navigate(generatePath(ROUTES.issueCredential.path))}
          type="primary"
        >
          {User === "verifier"
            ? REQUEST_FOR_VC_CREDS
            : User === "issuer"
            ? ISSUED_CREDENTIALS
            : REQUEST_FOR_VC}
        </Button>
      }
      title={REQUESTS}
    >
      <Space direction="vertical">
        <RequestsTable />
      </Space>
    </SiderLayoutContent>
  );
}
