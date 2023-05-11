import { Button, Col, Divider, Layout, Row, Space, Typography, notification } from "antd";
import { keccak256 } from "js-sha3";
import { ReactNode, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { z } from "zod";

import { ReactComponent as IconAlert } from "src/assets/icons/alert-triangle.svg";
import { ReactComponent as IconArrowLeft } from "src/assets/icons/arrow-narrow-left.svg";
import { ReactComponent as IconClose } from "src/assets/icons/x.svg";
import { useEnvContext } from "src/contexts/Env";
import { getStorageByKey, setStorageByKey } from "src/utils/browser";

const WARNING_ID = "warningNotification";

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
  // TODO - PID-684 reimplement warning notification conditionals to take into account authenticated & public routes instead once session flow is implemented.
  const { warningMessage } = useEnvContext();

  const warningKey = !!warningMessage && `${WARNING_ID}-${keccak256(warningMessage)}`;

  const [isShowingWarning, setShowWarning] = useState(
    !!warningKey && getStorageByKey({ defaultValue: true, key: warningKey, parser: z.boolean() })
  );

  const navigate = useNavigate();

  useEffect(() => {
    if (warningKey) {
      if (isShowingWarning) {
        notification.warning({
          closeIcon: <IconClose />,
          description: warningMessage,
          duration: 0,
          icon: <IconAlert />,
          key: warningKey,
          message: "Warning",
          onClose: () => setShowWarning(setStorageByKey({ key: warningKey, value: false })),
          placement: "bottom",
        });
      }

      Object.keys(localStorage).forEach((key) => {
        if (key.startsWith(WARNING_ID) && key !== warningKey) {
          localStorage.removeItem(key);
        }
      });
    }
  }, [warningMessage, isShowingWarning, warningKey]);

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
