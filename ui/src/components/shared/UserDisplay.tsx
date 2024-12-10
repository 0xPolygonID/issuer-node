import { Avatar, Card, Dropdown, Flex, Tag, Typography, theme } from "antd";
import { useNavigate } from "react-router-dom";
import { formatIdentifier } from "../../utils/forms";
import IconCheck from "src/assets/icons/check.svg?react";
import IconChevron from "src/assets/icons/chevron-selector-vertical.svg?react";
import IconCopy from "src/assets/icons/copy-01.svg?react";
import IconPlus from "src/assets/icons/plus.svg?react";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { ROUTES } from "src/routes";
import { isAsyncTaskDataAvailable } from "src/utils/async";
import { IDENTITY_ADD } from "src/utils/constants";

export function UserDisplay() {
  const { issuer } = useEnvContext();
  const { getSelectedIdentity, identifier, identityList, selectIdentity } = useIdentityContext();
  const navigate = useNavigate();
  const { token } = theme.useToken();
  const selectedIdentity = getSelectedIdentity();
  const selectedIdentityDisplayName = selectedIdentity ? selectedIdentity.displayName : "";

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
          onClick: () => selectIdentity(identity.identifier),
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
              ellipsis={{ tooltip: selectedIdentityDisplayName }}
              style={{ fontWeight: 600 }}
              type="secondary"
            >
              {selectedIdentityDisplayName}
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

              <Typography.Text
                copyable={{
                  icon: [<IconCopy key={0} />, <IconCheck key={1} />],
                  text: identifier,
                }}
                type="secondary"
              >
                {formatIdentifier(identifier, { short: true })}
              </Typography.Text>
            </Flex>
          </Flex>
        </Flex>
      </Card>
    </Dropdown>
  );
}
