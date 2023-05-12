import { Button, Grid, Layout, Row } from "antd";
import { useState } from "react";
import { Outlet } from "react-router-dom";

import { ReactComponent as IconMenu } from "src/assets/icons/menu-01.svg";
import { FeedbackLink } from "src/components/shared/FeedbackLink";
import { LogoLink } from "src/components/shared/LogoLink";
import { SiderMenu } from "src/components/shared/SiderMenu";
import { SIDER_WIDTH } from "src/utils/constants";

export function SiderLayout() {
  const [collapsed, setCollapsed] = useState(true);

  const { lg, md } = Grid.useBreakpoint();

  return (
    <Layout>
      {!lg && (
        <Layout.Header
          style={{
            borderBottom: "1px solid #EAECF0",
            height: 64,
            paddingLeft: md ? 32 : 16,
            paddingRight: md ? 16 : 8,
            position: "fixed",
            top: 0,
            width: "100%",
            zIndex: 3,
          }}
        >
          <Row align="middle" justify="space-between" style={{ height: "100%" }}>
            <LogoLink />

            <Button onClick={() => setCollapsed(!collapsed)} type="text">
              <IconMenu />
            </Button>
          </Row>
        </Layout.Header>
      )}

      <Layout.Sider
        collapsed={collapsed && !lg}
        collapsedWidth={0}
        collapsible
        style={{
          borderRight: "1px solid #EAECF0",
          height: "100vh",
          position: "fixed",
          zIndex: 2,
        }}
        trigger={null}
        width={SIDER_WIDTH}
      >
        <SiderMenu isBreakpoint={lg} onClick={() => setCollapsed(true)} />
      </Layout.Sider>

      <Layout.Content style={lg ? { marginLeft: SIDER_WIDTH } : { marginLeft: 0, marginTop: 64 }}>
        <Outlet />
      </Layout.Content>

      <FeedbackLink />

      {!collapsed && !lg && (
        <Row
          onClick={() => setCollapsed(true)}
          style={{
            background: "rgba(0,0,0,0.5)",
            cursor: "pointer",
            height: "100vh",
            position: "fixed",
            width: "100%",
          }}
        />
      )}
    </Layout>
  );
}
