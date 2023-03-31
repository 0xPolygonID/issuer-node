import { Dropdown, Row } from "antd";
import { generatePath, useNavigate } from "react-router-dom";

import { ReactComponent as IconDots } from "src/assets/icons/dots-vertical.svg";
import { ReactComponent as IconInfoCircle } from "src/assets/icons/info-circle.svg";
import { ReactComponent as IconTrash } from "src/assets/icons/trash-01.svg";
import { ROUTES } from "src/routes";

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

export function ConnectionsRowDropdown({ id }: { id: string }) {
  const navigate = useNavigate();
  const onMenuSelect = (key: string) => {
    if (key === "details") {
      navigate(generatePath(ROUTES.connectionDetails.path, { connectionID: id }));
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
