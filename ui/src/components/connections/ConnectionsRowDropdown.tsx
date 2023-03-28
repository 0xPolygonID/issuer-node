import { Dropdown, MenuProps, Row, message } from "antd";
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
    icon: <IconTrash />,
    key: "delete",
    label: "Delete connection",
  },
];

export function ConnectionsRowDropdown({ id }: { id: string }) {
  const navigate = useNavigate();
  const menuFunction: Record<"details" | "delete", () => Promise<void> | void> = {
    delete: () => void message.error("To develop"),
    details: () => navigate(generatePath(ROUTES.connectionDetails.path, { connectionID: id })),
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
