import { Button, Grid, Space, Tabs } from "antd";
import { ComponentType } from "react";
import { generatePath, useNavigate, useParams } from "react-router-dom";

import { RequestsTable } from "./RequestsTable";
import { ReactComponent as IconCreditCardPlus } from "src/assets/icons/credit-card-plus.svg";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { RequestsTabIDs } from "src/domain/request";
import { ROUTES } from "src/routes";
import { ISSUE_REQUEST, REQUEST, REQUEST_TABS } from "src/utils/constants";

const tabComponents: Record<RequestsTabIDs, ComponentType> = {
  Request: RequestsTable,
};

export function Request() {
  const navigate = useNavigate();
  const { tabID } = useParams();
  const { path } = ROUTES.request;

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

  return (
    <SiderLayoutContent
      description="Description for Request tab"
      extra={
        <Button
          icon={<IconCreditCardPlus />}
          onClick={() => navigate(generatePath(ROUTES.issueCredential.path))}
          type="primary"
        >
          {ISSUE_REQUEST}
        </Button>
      }
      title={REQUEST}
    >
      <Space direction="vertical">
        <Tabs
          activeKey={tabID}
          className={md ? undefined : "tab-responsive"}
          destroyInactiveTabPane
          items={REQUEST_TABS.map(({ id, tabID, title }) => {
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
