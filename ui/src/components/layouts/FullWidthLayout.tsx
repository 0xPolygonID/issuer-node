import { Col, Layout, Row } from "antd";
import { Outlet } from "react-router-dom";

import { LogoLink } from "src/components/shared/LogoLink";

export function FullWidthLayout({ background }: { background?: string }) {
  return (
    <Layout className={background} style={{ minHeight: "100vh" }}>
      <Layout.Header className={background} style={{ margin: "16px 0" }}>
        <Row align="middle">
          <Col style={{ width: 200 }}>
            <LogoLink />
          </Col>
        </Row>
      </Layout.Header>

      <Layout.Content>
        <Row align="middle" justify="center">
          <Col>
            <Outlet />
          </Col>
        </Row>
      </Layout.Content>
    </Layout>
  );
}
