import { Avatar, Dropdown, Flex, Row, Tooltip, Typography } from "antd";
import { useNavigate } from "react-router-dom";
import { formatIdentifier } from "../../utils/forms";
import IconCheck from "src/assets/icons/check.svg?react";
import IconChevron from "src/assets/icons/chevron-selector-vertical.svg?react";
import IconPlus from "src/assets/icons/plus.svg?react";
import { useEnvContext } from "src/contexts/Env";
import { useIssuerContext } from "src/contexts/Issuer";
import { ROUTES } from "src/routes";
import { isAsyncTaskDataAvailable } from "src/utils/async";
import { ISSUER_ADD } from "src/utils/constants";

export function UserDisplay() {
  const { issuer } = useEnvContext();
  const { handleChange, issuerDisplayName, issuerIdentifier, issuersList } = useIssuerContext();
  const navigate = useNavigate();

  const issuerItems = isAsyncTaskDataAvailable(issuersList)
    ? issuersList.data
        .toSorted((item) => (item.identifier === issuerIdentifier ? -1 : 0))
        .map(({ displayName, identifier }, index) => {
          const currentIssuer = identifier === issuerIdentifier;
          return {
            key: `${index}`,
            label: (
              <Flex
                align="center"
                className={currentIssuer ? "active" : ""}
                gap={16}
                justify="space-between"
              >
                <Flex
                  style={{
                    width: 242,
                  }}
                  vertical
                >
                  <Tooltip title={displayName}>
                    <Typography.Text ellipsis>{displayName}</Typography.Text>
                  </Tooltip>
                </Flex>
                {currentIssuer && <IconCheck />}
              </Flex>
            ),
            onClick: () => handleChange(identifier),
          };
        })
    : [];

  const items = [
    {
      key: "test",
      label: (
        <Flex gap={16}>
          <IconPlus style={{ height: 20, width: 20 }} />
          <Typography.Text>{ISSUER_ADD}</Typography.Text>
        </Flex>
      ),
      onClick: () => navigate(ROUTES.createIssuer.path),
    },
    ...issuerItems,
  ];

  return (
    <Flex gap={12} justify="space-between">
      <Flex>
        <Avatar shape="square" size="large" src={issuer.logo} />
      </Flex>

      <Flex style={{ maxWidth: 188, width: "100%" }} vertical>
        <Tooltip title={issuerDisplayName}>
          <Typography.Text ellipsis>{issuerDisplayName}</Typography.Text>
        </Tooltip>
        <Typography.Text type="secondary">{formatIdentifier(issuerIdentifier)}</Typography.Text>
      </Flex>

      <Flex>
        <Dropdown
          menu={{ items }}
          overlayClassName="issuers-dropdown"
          placement="bottom"
          trigger={["click"]}
        >
          <Row style={{ cursor: "pointer" }}>
            <IconChevron />
          </Row>
        </Dropdown>
      </Flex>
    </Flex>
  );
}
