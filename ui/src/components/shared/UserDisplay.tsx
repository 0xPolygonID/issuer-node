import { Avatar, Card, Dropdown, Flex, Tag, Tooltip, Typography, theme } from "antd";
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
  const { handleChange, identifier, identityDisplayName, identityList } = useIdentityContext();
  const navigate = useNavigate();
  const { token } = theme.useToken();

  const identityItems = isAsyncTaskDataAvailable(identityList)
    ? identityList.data.map((identity, index) => {
        const currentIdentity = identity.identifier === identifier;
        const formattedIdentifier = formatIdentifier(identity.identifier, { short: true });
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
                <Typography.Text
                  ellipsis={{ tooltip: identity.displayName || formattedIdentifier }}
                >
                  {identity.displayName || formattedIdentifier}
                </Typography.Text>
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
      key: "add",
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
    <Dropdown
      menu={{ items }}
      overlayClassName="identities-dropdown"
      placement="bottom"
      trigger={["click"]}
    >
      <Card className="user-display" styles={{ body: { cursor: "pointer", padding: 12 } }}>
        <Flex gap={12} vertical>
          <Flex justify="space-between">
            <Avatar shape="square" size="large" src={issuer.logo} />
            <IconChevron />
          </Flex>
          <Flex gap={4} vertical>
            <Typography.Text
              ellipsis={{ tooltip: identityDisplayName }}
              style={{ fontWeight: 600 }}
              type="secondary"
            >
              {identityDisplayName}
            </Typography.Text>

            <Flex align="center" gap={4}>
              <Tag
                style={{
                  background: "transparent",
                  border: "1px solid",
                  borderColor: token.colorInfoBorder,
                  borderRadius: 6,
                  color: token.colorTextSecondary,
                  fontSize: 12,
                  fontWeight: 500,
                  marginRight: 0,
                  padding: "3px 6px",
                }}
              >
                DID
              </Tag>
              <Tooltip title={identifier}>
                <Typography.Text type="secondary">
                  {formatIdentifier(identifier, { short: true })}
                </Typography.Text>
              </Tooltip>
            </Flex>
          </Flex>
        </Flex>
      </Card>
    </Dropdown>
  );
}
