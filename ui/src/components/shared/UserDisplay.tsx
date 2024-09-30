import { Avatar, Dropdown, Flex, Row, Tooltip, Typography } from "antd";
import { useNavigate } from "react-router-dom";
import { formatIdentifier } from "../../utils/forms";
import IconCheck from "src/assets/icons/check.svg?react";
import IconChevron from "src/assets/icons/chevron-selector-vertical.svg?react";
import IconPlus from "src/assets/icons/plus.svg?react";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { ROUTES } from "src/routes";
import { isAsyncTaskDataAvailable } from "src/utils/async";
import { IDENTITY_ADD } from "src/utils/constants";

export function UserDisplay() {
  const { issuer } = useEnvContext();
  const { handleChange, identifier, identitiesList, identityDisplayName } = useIdentityContext();
  const navigate = useNavigate();

  const identityItems = isAsyncTaskDataAvailable(identitiesList)
    ? identitiesList.data
        .toSorted((item) => (item.identifier === identifier ? -1 : 0))
        .map((identity, index) => {
          const currentIdentity = identity.identifier === identifier;
          return {
            key: `${index}`,
            label: (
              <Flex
                align="center"
                className={currentIdentity ? "active" : ""}
                gap={16}
                justify="space-between"
              >
                <Flex
                  style={{
                    width: 242,
                  }}
                  vertical
                >
                  <Tooltip title={identity.displayName}>
                    <Typography.Text ellipsis>{identity.displayName}</Typography.Text>
                  </Tooltip>
                </Flex>
                {currentIdentity && <IconCheck />}
              </Flex>
            ),
            onClick: () => handleChange(identity.identifier),
          };
        })
    : [];

  const items = [
    {
      key: "test",
      label: (
        <Flex gap={16}>
          <IconPlus style={{ height: 20, width: 20 }} />
          <Typography.Text>{IDENTITY_ADD}</Typography.Text>
        </Flex>
      ),
      onClick: () => navigate(ROUTES.createIdentity.path),
    },
    ...identityItems,
  ];

  return (
    <Flex gap={12} justify="space-between">
      <Flex>
        <Avatar shape="square" size="large" src={issuer.logo} />
      </Flex>

      <Flex style={{ maxWidth: 188, width: "100%" }} vertical>
        <Tooltip title={identityDisplayName}>
          <Typography.Text ellipsis>{identityDisplayName}</Typography.Text>
        </Tooltip>
        <Typography.Text type="secondary">{formatIdentifier(identifier)}</Typography.Text>
      </Flex>

      <Flex>
        <Dropdown
          menu={{ items }}
          overlayClassName="identities-dropdown"
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
