import { Layout } from "antd";
import { Outlet } from "react-router-dom";

import { FeedbackLink } from "src/components/shared/FeedbackLink";
import { SiderMenu } from "src/components/shared/SiderMenu";
import { SIDER_WIDTH } from "src/utils/constants";

export function SiderLayout() {
  return (
    <Layout>
      <Layout.Sider
        style={{
          borderRight: "1px solid #EAECF0",
          height: "100vh",
          padding: "32px 24px",
          position: "fixed",
        }}
        width={SIDER_WIDTH}
      >
        <SiderMenu />
      </Layout.Sider>

      <Layout className="bg-light" style={{ marginLeft: SIDER_WIDTH, minHeight: "100vh" }}>
        <Outlet />
      </Layout>

      <FeedbackLink />
    </Layout>
  );
}
