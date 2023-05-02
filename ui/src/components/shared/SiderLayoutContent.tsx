import { Button, Col, Divider, Layout, Row, Space, Typography } from "antd";
import { ReactNode } from "react";
import { useNavigate } from "react-router-dom";

import { ReactComponent as IconArrowLeft } from "src/assets/icons/arrow-narrow-left.svg";

export function SiderLayoutContent({
  children,
  description,
  extra = null,
  showBackButton,
  showDivider = false,
  title,
}: {
  children: ReactNode;
  description?: string;
  extra?: ReactNode;
  showBackButton?: boolean;
  showDivider?: boolean;
  title?: string;
}) {
  const navigate = useNavigate();
  return (
    <>
      <Layout.Header
        className="bg-light"
        style={{ height: "auto", padding: 32, paddingBottom: showDivider ? 0 : 12 }}
      >
        <Row justify="space-between">
          <Space align="start" size="large">
            {showBackButton && (
              <Button
                icon={<IconArrowLeft style={{ marginRight: 0 }} />}
                onClick={() => navigate(-1)}
              />
            )}

            <Col style={{ lineHeight: "1rem", maxWidth: 585 }}>
              <Typography.Title level={3}>{title}</Typography.Title>

              {description && <Typography.Text type="secondary">{description}</Typography.Text>}
            </Col>
          </Space>

          {extra}
        </Row>
      </Layout.Header>

      {showDivider && <Divider />}

      <Layout.Content style={{ padding: 32, paddingBottom: 64, paddingTop: 0 }}>
        {children}
      </Layout.Content>
    </>
  );
}
