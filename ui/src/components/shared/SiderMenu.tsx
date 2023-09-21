import { Col, Divider, Menu, Row, Space, Tag, Typography } from "antd";
import { generatePath, matchRoutes, useLocation, useNavigate } from "react-router-dom";
import { ReactComponent as IconNotification } from "src/assets/icons/bell-notification.svg";
import { ReactComponent as IconCredentials } from "src/assets/icons/credit-card-refresh.svg";
import { ReactComponent as IconLogout } from "src/assets/icons/logout-user.svg";
import { ReactComponent as IconIssuerState } from "src/assets/icons/switch-horizontal.svg";
import { ReactComponent as IconConnections } from "src/assets/icons/users-01.svg";
import { UserDisplay } from "src/components/shared/UserDisplay";
import { useIssuerStateContext } from "src/contexts/IssuerState";
import { ROUTES } from "src/routes";
import { isAsyncTaskDataAvailable } from "src/utils/async";
import {
  CONNECTIONS,
  CREDENTIALS,
  CREDENTIALS_TABS,
  ISSUER_STATE,
  NOTIFICATION,
} from "src/utils/constants";

export function SiderMenu({
  isBreakpoint,
  onClick,
}: {
  isBreakpoint?: boolean;
  onClick: () => void;
}) {
  // const { buildTag } = useEnvContext();
  const { status } = useIssuerStateContext();
  const { pathname } = useLocation();
  const navigate = useNavigate();

  const connectionsPath = ROUTES.connections.path;
  const credentialsPath = ROUTES.credentials.path;
  const issuerStatePath = ROUTES.issuerState.path;
  const notificationPath = ROUTES.notification.path;

  const getSelectedKey = (): string[] => {
    if (
      matchRoutes(
        [
          { path: credentialsPath },
          { path: ROUTES.credentialDetails.path },
          { path: ROUTES.issueCredential.path },
          { path: ROUTES.linkDetails.path },
        ],
        pathname
      )
    ) {
      return [credentialsPath];
    } else if (
      matchRoutes([{ path: connectionsPath }, { path: ROUTES.connectionDetails.path }], pathname)
    ) {
      return [connectionsPath];
    } else if (matchRoutes([{ path: issuerStatePath }], pathname)) {
      return [issuerStatePath];
    } else if (matchRoutes([{ path: notificationPath }], pathname)) {
      return [notificationPath];
    }

    return [];
  };

  const onMenuClick = (path: string) => {
    onClick();
    navigate(path);
  };

  return (
    <Row
      className="menu-sider-layout"
      justify="space-between"
      style={{
        padding: isBreakpoint ? "32px 24px" : "96px 24px 32px",
      }}
    >
      <Col>
        <UserDisplay />

        <Divider />

        <Menu
          items={[
            {
              icon: <IconCredentials />,
              key: credentialsPath,
              label: CREDENTIALS,
              onClick: () =>
                onMenuClick(
                  generatePath(credentialsPath, {
                    tabID: CREDENTIALS_TABS[0]?.tabID,
                  })
                ),
              title: "",
            },
            {
              icon: <IconConnections />,
              key: connectionsPath,
              label: CONNECTIONS,
              onClick: () => onMenuClick(connectionsPath),
              title: "",
            },
            {
              icon: <IconIssuerState />,

              key: issuerStatePath,

              label:
                isAsyncTaskDataAvailable(status) && status.data ? (
                  <Space>
                    {ISSUER_STATE}
                    <Tag color="purple" style={{ fontSize: 12 }}>
                      Pending actions
                    </Tag>
                  </Space>
                ) : (
                  ISSUER_STATE
                ),
              onClick: () => onMenuClick(issuerStatePath),
              title: "",
            },
            {
              icon: <IconNotification />,
              key: notificationPath,
              label: NOTIFICATION,
              onClick: () => onMenuClick(notificationPath),
              title: "",
            },
          ]}
          selectedKeys={getSelectedKey()}
        />
      </Col>

      <Space direction="vertical" size={10}>
        <Menu
          items={[
            {
              icon: <IconLogout />,
              key: "documentation",
              label: (
                <Typography.Link>
                  <Row justify="space-between">
                    <span>Sign Out</span>
                  </Row>
                </Typography.Link>
              ),
            },
          ]}
        />
        {/* {isBreakpoint && (

          <Space>
            <LogoLink />
            {buildTag && <Tag>{buildTag}</Tag>}
          </Space>
        )} */}
      </Space>
    </Row>
  );
}
