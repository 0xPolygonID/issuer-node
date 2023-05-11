import { Button, Grid, Space, Tabs } from "antd";
import { ComponentType } from "react";
import { Navigate, generatePath, useNavigate, useParams } from "react-router-dom";

import { ReactComponent as IconCreditCardPlus } from "src/assets/icons/credit-card-plus.svg";
import { CredentialsTable } from "src/components/credentials/CredentialsTable";
import { LinksTable } from "src/components/credentials/LinksTable";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { CredentialsTabIDs } from "src/domain";
import { ROUTES } from "src/routes";
import { CREDENTIALS, CREDENTIALS_TABS, ISSUE_CREDENTIAL } from "src/utils/constants";

const tabComponents: Record<CredentialsTabIDs, ComponentType> = {
  issued: CredentialsTable,
  links: LinksTable,
};

export function Credentials() {
  const navigate = useNavigate();
  const { tabID } = useParams();
  const { path } = ROUTES.credentials;

  const { md } = Grid.useBreakpoint();

  const goToTab = (key: string): void => {
    if (key !== tabID) {
      navigate(
        generatePath(path, {
          tabID: key,
        })
      );
    }
  };

  if (!CREDENTIALS_TABS.some((tabs) => tabs.tabID === tabID)) {
    return (
      <Navigate
        to={generatePath(path, {
          tabID: CREDENTIALS_TABS[0].tabID,
        })}
      />
    );
  }

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
        <Tabs
          activeKey={tabID}
          className={md ? undefined : "tab-responsive"}
          destroyInactiveTabPane
          items={CREDENTIALS_TABS.map(({ id, tabID, title }) => {
            const Component = tabComponents[id];

            return {
              children: <Component />,
              key: tabID,
              label: title,
            };
          })}
          onTabClick={goToTab}
        />
      </Space>
    </SiderLayoutContent>
  );
}
