import { Col, Divider, Menu, Row, Space, Tag, Typography } from "antd";
import { generatePath, matchRoutes, useLocation, useNavigate } from "react-router-dom";

import { ReactComponent as IconCredentials } from "src/assets/icons/credit-card-refresh.svg";
import { ReactComponent as IconFile } from "src/assets/icons/file-05.svg";
import { ReactComponent as IconSchema } from "src/assets/icons/file-search-02.svg";
import { ReactComponent as IconLink } from "src/assets/icons/link-external-01.svg";
import { ReactComponent as IconIssuerState } from "src/assets/icons/switch-horizontal.svg";
import { ReactComponent as IconConnections } from "src/assets/icons/users-01.svg";
import { LogoLink } from "src/components/shared/LogoLink";
import { UserDisplay } from "src/components/shared/UserDisplay";
import { useEnvContext } from "src/contexts/Env";
import { useIssuerStateContext } from "src/contexts/IssuerState";
import { ROUTES } from "src/routes";
import { isAsyncTaskDataAvailable } from "src/utils/async";
import {
  CONNECTIONS,
  CREDENTIALS,
  CREDENTIALS_TABS,
  ISSUER_STATE,
  SCHEMAS,
  TUTORIALS_URL,
} from "src/utils/constants";

export function SiderMenu({
  isBreakpoint,
  onClick,
}: {
  isBreakpoint?: boolean;
  onClick: () => void;
}) {
  const { buildTag } = useEnvContext();
  const { status } = useIssuerStateContext();
  const { pathname } = useLocation();
  const navigate = useNavigate();

  const connectionsPath = ROUTES.connections.path;
  const credentialsPath = ROUTES.credentials.path;
  const issuerStatePath = ROUTES.issuerState.path;
  const schemasPath = ROUTES.schemas.path;

  const getSelectedKey = (): string[] => {
    if (
      matchRoutes(
        [
          { path: schemasPath },
          { path: ROUTES.importSchema.path },
          { path: ROUTES.schemaDetails.path },
        ],
        pathname
      )
    ) {
      return [schemasPath];
    } else if (
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
              icon: <IconSchema />,
              key: schemasPath,
              label: SCHEMAS,
              onClick: () => onMenuClick(schemasPath),
              title: "",
            },
            {
              icon: <IconCredentials />,
              key: credentialsPath,
              label: CREDENTIALS,
              onClick: () =>
                onMenuClick(
                  generatePath(credentialsPath, {
                    tabID: CREDENTIALS_TABS[0].tabID,
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
          ]}
          selectedKeys={getSelectedKey()}
        />
      </Col>

      <Space direction="vertical" size="large">
        <Menu
          items={[
            {
              icon: <IconFile />,
              key: "documentation",
              label: (
                <Typography.Link href={TUTORIALS_URL} target="_blank">
                  <Row justify="space-between">
                    <span>Documentation</span>

                    <IconLink className="icon-secondary" height={16} />
                  </Row>
                </Typography.Link>
              ),
            },
          ]}
        />
        {isBreakpoint && (
          <Space>
            <LogoLink />

            {buildTag && <Tag>{buildTag}</Tag>}
          </Space>
        )}
      </Space>
    </Row>
  );
}
