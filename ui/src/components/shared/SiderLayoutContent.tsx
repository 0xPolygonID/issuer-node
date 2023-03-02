import { Button, Col, Divider, Layout, Row, Space, Typography } from "antd";
import { ReactNode } from "react";
import { Link } from "react-router-dom";

import { ReactComponent as IconArrowLeft } from "src/assets/icons/arrow-narrow-left.svg";
import { CONTENT_WIDTH } from "src/utils/constants";

export function SiderLayoutContent({
  backButtonLink,
  children,
  description,
  extra = null,
  showDivider = false,
  title,
}: {
  backButtonLink?: string;
  children: ReactNode;
  description?: string;
  extra?: ReactNode;
  showDivider?: boolean;
  title?: string;
}) {
  return (
    <>
      <Layout.Header
        className="bg-light"
        style={{ height: "auto", padding: 32, paddingBottom: showDivider ? 0 : 12 }}
      >
        <Row justify="space-between">
          <Space align="start" size="large">
            {backButtonLink && (
              <Link to={backButtonLink}>
                <Button icon={<IconArrowLeft style={{ marginRight: 0 }} />} />
              </Link>
            )}

            <Col style={{ lineHeight: "1rem", maxWidth: CONTENT_WIDTH }}>
              <Typography.Title level={3}>{title}</Typography.Title>

              {description && <Typography.Text type="secondary">{description}</Typography.Text>}
            </Col>
          </Space>

          {extra}
        </Row>
      </Layout.Header>

      {showDivider && <Divider />}

      <Layout.Content style={{ padding: 32, paddingTop: 0 }}>{children}</Layout.Content>
    </>
  );
}
