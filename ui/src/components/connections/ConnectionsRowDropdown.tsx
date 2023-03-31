import { Dropdown, Row, message } from "antd";

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
  const onMenuSelect = (key: string) => {
    if (key === "details") {
      void message.error("Details not yet implemented");
    } else if (key === "delete") {
      void message.error("Delete not yet implemented");
    }
  };

  return (
    <Dropdown
      menu={{
        items: MENU_ITEMS,
        onClick: ({ key }) => {
          onMenuSelect(key);
        },
      }}
    >
      <Row>
        <IconDots className="icon-secondary" />
      </Row>
    </Dropdown>
  );
}
