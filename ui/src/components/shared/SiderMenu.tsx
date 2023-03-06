import { Col, Divider, Menu, Row, Space, Tag, Typography } from "antd";
import { generatePath, matchPath, useLocation, useNavigate } from "react-router-dom";

import { ReactComponent as IconFile } from "src/assets/icons/file-05.svg";
import { ReactComponent as IconSchema } from "src/assets/icons/file-search-02.svg";
import { ReactComponent as IconLink } from "src/assets/icons/link-external-01.svg";
import { LogoLink } from "src/components/shared/LogoLink";
import { UserDropdown } from "src/components/shared/UserDropdown";
import { ROUTES } from "src/routes";
import { SCHEMAS_TABS, TUTORIALS_URL } from "src/utils/constants";

export function SiderMenu() {
  const { pathname } = useLocation();
  const navigate = useNavigate();

  const pathSchemas = ROUTES.schemas.path;
  const pathIssueClaim = ROUTES.issueClaim.path;

  const getSelectedKey = (): string[] => {
    if (matchPath(pathSchemas, pathname) || matchPath(pathIssueClaim, pathname)) {
      return ["schemas"];
    } else {
      return [];
    }
  };

  return (
    <Row className="menu-sider-layout" justify="space-between">
      <Col>
        <UserDropdown />

        <Divider />

        <Menu
          items={[
            {
              icon: <IconSchema />,
              key: "schemas",
              label: "Schemas",
              onClick: () =>
                navigate(
                  generatePath(pathSchemas, {
                    tabID: SCHEMAS_TABS[0].tabID,
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
