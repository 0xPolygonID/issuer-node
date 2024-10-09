import { Button, Divider, Space, message } from "antd";
import { useCallback, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import IconPlus from "src/assets/icons/plus.svg?react";
import { IdentitiesTable } from "src/components/identities/IdentitiesTable";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useIdentityContext } from "src/contexts/Identity";
import { ROUTES } from "src/routes";
import { makeRequestAbortable } from "src/utils/browser";
import { IDENTITIES, IDENTITY_ADD } from "src/utils/constants";

export function Identities() {
  const { fetchIdentities } = useIdentityContext();

  const [, messageContext] = message.useMessage();
  const navigate = useNavigate();

  const fetchData = useCallback(() => {
    const { aborter } = makeRequestAbortable(fetchIdentities);
    return aborter;
  }, [fetchIdentities]);

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
            onClick={() => navigate(ROUTES.createIdentity.path)}
            type="primary"
          >
            {IDENTITY_ADD}
          </Button>
        }
        title={IDENTITIES}
      >
        <Divider />
        <Space direction="vertical" size="large">
          <IdentitiesTable handleAddIdentity={() => navigate(ROUTES.createIdentity.path)} />
        </Space>
      </SiderLayoutContent>
    </>
  );
}
