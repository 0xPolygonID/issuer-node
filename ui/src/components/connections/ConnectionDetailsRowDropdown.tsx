import { Dropdown, Row } from "antd";

import { ReactComponent as IconDots } from "src/assets/icons/dots-vertical.svg";
import { ReactComponent as IconInfoCircle } from "src/assets/icons/info-circle.svg";
import { ReactComponent as IconTrash } from "src/assets/icons/trash-01.svg";
import { ReactComponent as IconClose } from "src/assets/icons/x.svg";

const MENU_ITEMS = [
  {
    icon: <IconInfoCircle />,
    key: "details",
    label: "Details",
  },
  {
    key: "divider1",
    type: "divider",
  },
  {
    danger: true,
    icon: <IconClose />,
    key: "revoke",
    label: "Revoke",
  },
  {
    key: "divider2",
    type: "divider",
  },
  {
    danger: true,
    icon: <IconTrash />,
    key: "delete",
    label: "Delete",
  },
];

export function ConnectionDetailsRowDropdown() {
  return (
    <Dropdown menu={{ items: MENU_ITEMS }}>
      <Row>
        <IconDots className="icon-secondary" />
      </Row>
    </Dropdown>
  );
}
