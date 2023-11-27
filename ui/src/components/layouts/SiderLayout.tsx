import { Button, Grid, Layout, Row } from "antd";
import { useState } from "react";
import { Outlet } from "react-router-dom";

import IconMenu from "src/assets/icons/menu-01.svg?react";
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
          className="responsive-header"
          style={{
            paddingLeft: md ? 32 : 16,
            paddingRight: md ? 16 : 8,
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
        className="sider-layout"
        collapsed={collapsed && !lg}
        collapsedWidth={0}
        collapsible
        trigger={null}
        width={SIDER_WIDTH}
      >
        <SiderMenu isBreakpoint={lg} onClick={() => setCollapsed(true)} />
      </Layout.Sider>

      <Layout.Content style={lg ? { marginLeft: SIDER_WIDTH } : { marginLeft: 0, marginTop: 64 }}>
        <Outlet />
      </Layout.Content>

      <FeedbackLink />

      {!collapsed && !lg && <Row className="background-sider" onClick={() => setCollapsed(true)} />}
    </Layout>
  );
}
