import { Button, Space, Tabs } from "antd";
import { useState } from "react";
import { Navigate, generatePath, useNavigate, useParams } from "react-router-dom";

import { ReactComponent as IconUpload } from "src/assets/icons/upload-01.svg";
import { MySchemas } from "src/components/schemas/MySchemas";
import { Explainer } from "src/components/shared/Explainer";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { ROUTES } from "src/routes";
import { SCHEMAS_TABS, TUTORIALS_URL } from "src/utils/constants";

export function Schemas() {
  // TODO - this will be handled via the API, see CID-155
  const [isExplainerShowing, setExplainerShowing] = useState(true);

  const navigate = useNavigate();
  const { tabID } = useParams();
  const { path } = ROUTES.schemas;

  const goToTab = (key: string): void => {
    if (key !== tabID) {
      navigate(
        generatePath(path, {
          tabID: key,
        })
      );
    }
  };

  if (!SCHEMAS_TABS.some((tabs) => tabs.tabID === tabID)) {
    return (
      <Navigate
        to={generatePath(path, {
          tabID: SCHEMAS_TABS[0].tabID,
        })}
      />
    );
  }

  return (
    <SiderLayoutContent
      description="Verifiable credential schemas help to ensure the structure and data formatting across different services."
      extra={
        <Space align="start" size="middle">
          <Button
            icon={<IconUpload />}
            onClick={() => navigate(ROUTES.importSchema.path)}
            type="primary"
          >
            Import schema
          </Button>
        </Space>
      }
      title="Schemas"
    >
      <Space direction="vertical">
        {isExplainerShowing && (
          <Explainer
            CTA={{ label: "Learn more", url: TUTORIALS_URL }}
            description="Learn about schema types, attributes, naming conventions, data types and more."
            handleDismiss={() => setExplainerShowing(false)}
            image="/images/illustration-explainer.svg"
            title="Credential schemas explained"
          />
        )}
        <Tabs
          activeKey={tabID}
          destroyInactiveTabPane
          items={SCHEMAS_TABS.map(({ id, tabID, title }) => {
            return {
              children: <MySchemas showActive={id === "mySchemas"} />,
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
