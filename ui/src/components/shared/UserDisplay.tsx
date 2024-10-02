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
  const { handleChange, identifier, identitiesList, identityDisplayName } = useIdentityContext();
  const navigate = useNavigate();
  const { token } = theme.useToken();

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
            <Tooltip title={identityDisplayName}>
              <Typography.Text ellipsis style={{ fontWeight: 600 }} type="secondary">
                {identityDisplayName}
              </Typography.Text>
            </Tooltip>

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
                  {formatIdentifier(identifier, true)}
                </Typography.Text>
              </Tooltip>
            </Flex>
          </Flex>
        </Flex>
      </Card>
    </Dropdown>

    // <Flex gap={12} justify="space-between">
    //   <Flex>
    //     <Avatar shape="square" size="large" src={issuer.logo} />
    //   </Flex>

    //   <Flex style={{ maxWidth: 188, width: "100%" }} vertical>
    //     <Tooltip title={identityDisplayName}>
    //       <Typography.Text ellipsis>{identityDisplayName}</Typography.Text>
    //     </Tooltip>
    //     <Typography.Text type="secondary">{formatIdentifier(identifier)}</Typography.Text>
    //   </Flex>

    //   <Flex>
    //     <Dropdown
    //       menu={{ items }}
    //       overlayClassName="identities-dropdown"
    //       placement="bottom"
    //       trigger={["click"]}
    //     >
    //       <Row style={{ cursor: "pointer" }}>
    //         <IconChevron />
    //       </Row>
    //     </Dropdown>
    //   </Flex>
    // </Flex>
  );
}
