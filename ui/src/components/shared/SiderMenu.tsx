import { Col, Divider, Menu, Row, Space, Tag, Typography } from "antd";
import { generatePath, matchPath, matchRoutes, useLocation, useNavigate } from "react-router-dom";

import { ReactComponent as IconCredentials } from "src/assets/icons/credit-card-refresh.svg";
import { ReactComponent as IconFile } from "src/assets/icons/file-05.svg";
import { ReactComponent as IconSchema } from "src/assets/icons/file-search-02.svg";
import { ReactComponent as IconLink } from "src/assets/icons/link-external-01.svg";
import { LogoLink } from "src/components/shared/LogoLink";
import { UserDisplay } from "src/components/shared/UserDisplay";
import { ROUTES } from "src/routes";
import { CREDENTIALS, CREDENTIALS_TABS, SCHEMAS, TUTORIALS_URL } from "src/utils/constants";

export function SiderMenu() {
  const { pathname } = useLocation();
  const navigate = useNavigate();

  const credentialsPath = ROUTES.credentials.path;
  const schemasPath = ROUTES.schemas.path;

  const getSelectedKey = (): string[] => {
    if (matchPath(schemasPath, pathname)) {
      return ["schemas"];
    } else if (
      matchRoutes([{ path: credentialsPath }, { path: ROUTES.issueCredential.path }], pathname)
    ) {
      return ["credentials"];
    }

    return [];
  };

  return (
    <Row className="menu-sider-layout" justify="space-between">
      <Col>
        <UserDisplay />

        <Divider />

        <Menu
          items={[
            {
              icon: <IconSchema />,
              // TODO - these keys need to be typed.
              key: "schemas",
              label: SCHEMAS,
              onClick: () => navigate(schemasPath),
            },
            {
              icon: <IconCredentials />,
              key: "credentials",
              label: CREDENTIALS,
              onClick: () =>
                navigate(
                  generatePath(credentialsPath, {
                    tabID: CREDENTIALS_TABS[0].tabID,
                  })
                ),
            },
          ]}
          selectedKeys={getSelectedKey()}
        />
      </Col>

      <Col>
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
          selectedKeys={getSelectedKey()}
        />

        <Space style={{ marginTop: 42 }}>
          <LogoLink />

          <Tag>Testnet</Tag>
        </Space>
      </Col>
    </Row>
  );
}
