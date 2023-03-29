import { Col, Divider, Menu, Row, Space, Typography } from "antd";
import { generatePath, matchRoutes, useLocation, useNavigate } from "react-router-dom";

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
      matchRoutes([{ path: credentialsPath }, { path: ROUTES.issueCredential.path }], pathname)
    ) {
      return [credentialsPath];
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
              key: schemasPath,
              label: SCHEMAS,
              onClick: () => navigate(schemasPath),
            },
            {
              icon: <IconCredentials />,
              key: credentialsPath,
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
        />

        <Space style={{ marginTop: 40 }}>
          <LogoLink />
        </Space>
      </Col>
    </Row>
  );
}
