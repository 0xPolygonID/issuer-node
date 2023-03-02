import { Dropdown, MenuProps, Row, message } from "antd";

import { schemasUpdate } from "src/adapters/api/schemas";
import { ReactComponent as IconArchive } from "src/assets/icons/archive.svg";
import { ReactComponent as IconDots } from "src/assets/icons/dots-vertical.svg";
import { useAuthContext } from "src/hooks/useAuthContext";

const MENU_ITEMS = [
  {
    icon: <IconArchive />,
    key: "archive",
    label: "Move to archive",
  },
];

export function SchemaRowDropdown({ id, onAction }: { id: string; onAction: () => void }) {
  const { account, authToken } = useAuthContext();
  const menuFunction: Record<"archive", () => Promise<void> | void> = {
    archive: async () => {
      if (authToken && account?.organization) {
        const isUpdated = await schemasUpdate({
          issuerID: account.organization,
          payload: { active: false },
          schemaID: id,
          token: authToken,
        });

        if (isUpdated) {
          void message.success("Claim schema moved to archive.");
          onAction();
        }
      }
    },
  };

  const onMenuSelect: MenuProps["onClick"] = ({ domEvent, key }) => {
    domEvent.stopPropagation();
    if (key === "archive") {
      void menuFunction[key]();
    }
  };

  return (
    <Dropdown menu={{ items: MENU_ITEMS, onClick: onMenuSelect }}>
      <Row>
        <IconDots className="icon-secondary" />
      </Row>
    </Dropdown>
  );
}
