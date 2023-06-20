import { Button, Col, Divider, Grid, Layout, Row, Typography, notification } from "antd";
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
  description?: ReactNode;
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

  const { md } = Grid.useBreakpoint();

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
    <Layout className="bg-light" style={{ minHeight: "100vh" }}>
      <Row
        gutter={[0, 16]}
        justify="space-between"
        style={{
          height: "auto",
          padding: md
            ? `32px 32px ${showDivider ? 0 : "12px"} `
            : `16px 16px ${showDivider ? 0 : "12px"} `,
        }}
      >
        <Row gutter={[24, 16]}>
          {showBackButton && (
            <Col>
              <Button
                icon={<IconArrowLeft style={{ marginRight: 0 }} />}
                onClick={() => navigate(-1)}
              />
            </Col>
          )}

          <Col style={{ lineHeight: "1rem", maxWidth: 585 }}>
            <Typography.Title level={3}>{title}</Typography.Title>

            {description && <Typography.Text type="secondary">{description}</Typography.Text>}
          </Col>
        </Row>

        {extra}
      </Row>

      {showDivider && <Divider />}

      <Row style={{ padding: `0 ${md ? "32px" : 0} 64px` }}>{children}</Row>
    </Layout>
  );
}
