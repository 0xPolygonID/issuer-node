import { Col, Divider, Menu, Modal, Row, Space, Typography } from "antd";
import { useState } from "react";
import { generatePath, matchRoutes, useLocation, useNavigate } from "react-router-dom";
import { ReactComponent as IconNotification } from "src/assets/icons/bell-notification.svg";
import { ReactComponent as IconCredentials } from "src/assets/icons/credit-card-refresh.svg";
import { ReactComponent as IconLogout } from "src/assets/icons/logout-user.svg";
import { ReactComponent as IconProfile } from "src/assets/icons/profile.svg";
import { ReactComponent as IconRequest } from "src/assets/icons/switch-horizontal.svg";
import { ReactComponent as IconConnections } from "src/assets/icons/users-01.svg";
import { UserDisplay } from "src/components/shared/UserDisplay";
import { ROUTES } from "src/routes";
import {
  ALL_REQUEST,
  CONNECTIONS,
  CREDENTIALS,
  CREDENTIALS_TABS,
  NOTIFICATION,
  PROFILE,
} from "src/utils/constants";

export function SiderMenu({
  isBreakpoint,
  onClick,
}: {
  isBreakpoint?: boolean;
  onClick: () => void;
}) {
  const { pathname } = useLocation();
  const navigate = useNavigate();
  const [status, setStatus] = useState<boolean>(false);

  const connectionsPath = ROUTES.connections.path;
  const credentialsPath = ROUTES.credentials.path;
  const profilepath = ROUTES.profile.path;
  const issuerStatePath = ROUTES.issuerState.path;
  const notificationPath = ROUTES.notification.path;
  const loginPath = ROUTES.login.path;
  const requestPath = ROUTES.request.path;
  const User = localStorage.getItem("user");

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
    } else if (matchRoutes([{ path: profilepath }], pathname)) {
      return [profilepath];
    } else if (matchRoutes([{ path: requestPath }], pathname)) {
      return [requestPath];
    }

    return [];
  };
  const profileStatus = localStorage.getItem("profile");
  const onMenuClick = (path: string) => {
    if (profileStatus === "true") {
      onClick();
      navigate(path);
    } else {
      setStatus(true);
    }
  };
  const onLogout = (path: string) => {
    localStorage.clear();
    navigate(path);
  };
  const handleOk = () => {
    setStatus(false);
  };

  return (
    <>
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

          {User !== "verifier" && User !== "issuer" ? (
            <Menu
              items={[
                {
                  icon: <IconProfile />,
                  key: profilepath,
                  label: PROFILE,
                  onClick: () => onMenuClick(profilepath),
                  title: "",
                },
                {
                  icon: <IconRequest />,
                  key: requestPath,
                  label: ALL_REQUEST,
                  onClick: () => onMenuClick(requestPath),
                  title: "",
                },
                {
                  icon: <IconNotification />,
                  key: notificationPath,
                  label: NOTIFICATION,
                  onClick: () => onMenuClick(notificationPath),
                  title: "",
                },
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
              ]}
              selectedKeys={getSelectedKey()}
            />
          ) : (
            <Menu
              items={[
                {
                  icon: <IconRequest />,
                  key: requestPath,
                  label: ALL_REQUEST,
                  onClick: () => onMenuClick(requestPath),
                  title: "",
                },
                {
                  icon: <IconNotification />,
                  key: notificationPath,
                  label: NOTIFICATION,
                  onClick: () => onMenuClick(notificationPath),
                  title: "",
                },
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
              ]}
              selectedKeys={getSelectedKey()}
            />
          )}
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
                onClick: () => onLogout(loginPath),
              },
            ]}
          />
        </Space>
      </Row>
      <Modal onCancel={handleOk} onOk={handleOk} open={status} title="Important">
        <p>Please update profile first</p>
      </Modal>
    </>
  );
}
