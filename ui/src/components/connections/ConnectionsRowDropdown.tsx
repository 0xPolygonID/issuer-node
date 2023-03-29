import { Dropdown, MenuProps, Row, message } from "antd";

import { ReactComponent as IconDots } from "src/assets/icons/dots-vertical.svg";
import { ReactComponent as IconInfoCircle } from "src/assets/icons/info-circle.svg";
import { ReactComponent as IconTrash } from "src/assets/icons/trash-01.svg";

const MENU_ITEMS = [
  {
    icon: <IconInfoCircle />,
    key: "details",
    label: "Details",
  },
  {
    key: "divider",
    type: "divider",
  },
  {
    danger: true,
    icon: <IconTrash />,
    key: "delete",
    label: "Delete connection",
  },
];

export function ConnectionsRowDropdown() {
  const menuFunction: Record<"details" | "delete", () => Promise<void> | void> = {
    delete: () => void message.error("To develop"),
    details: () => void message.error("To develop"),
  };

  const onMenuSelect: MenuProps["onClick"] = ({ domEvent, key }) => {
    domEvent.stopPropagation();
    if (key === "details" || key === "delete") {
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
