import { Button, Divider, Space, message } from "antd";
import { useCallback, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import IconPlus from "src/assets/icons/plus.svg?react";
import { IssuersTable } from "src/components/issuers/IssuersTable";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useIssuerContext } from "src/contexts/Issuer";
import { ROUTES } from "src/routes";
import { makeRequestAbortable } from "src/utils/browser";
import { ISSUERS, ISSUER_ADD } from "src/utils/constants";

export function Issuers() {
  const { fetchIssuers } = useIssuerContext();

  const [, messageContext] = message.useMessage();
  const navigate = useNavigate();

  const fetchData = useCallback(() => {
    const { aborter } = makeRequestAbortable(fetchIssuers);
    return aborter;
  }, [fetchIssuers]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  return (
    <>
      {messageContext}

      <SiderLayoutContent
        description="Add new identities and view existing identity details"
        extra={
          <Button
            icon={<IconPlus />}
            onClick={() => navigate(ROUTES.createIssuer.path)}
            type="primary"
          >
            {ISSUER_ADD}
          </Button>
        }
        title={ISSUERS}
      >
        <Divider />
        <Space direction="vertical" size="large">
          <IssuersTable handleAddIssuer={() => navigate(ROUTES.createIssuer.path)} />
        </Space>
      </SiderLayoutContent>
    </>
  );
}
