import { Row, Tooltip, Typography } from "antd";
import { ReactNode } from "react";

import { ReactComponent as ChevronRightIcon } from "src/assets/icons/chevron-right.svg";
import { ObjectAttribute } from "src/domain";

export function AttributeBreadcrumb({ parents }: { parents: ObjectAttribute[] }) {
  const last = parents.length > 0 ? parents[parents.length - 1] : undefined;
  const rest = parents.slice(0, -1);
  return last && rest.length ? (
    <Row align="bottom">
      <Tooltip
        placement="topLeft"
        title={
          <Row align="bottom">
            {rest.reduce(
              (acc: ReactNode[], curr: ObjectAttribute, index): ReactNode[] => [
                ...acc,
                acc.length > 0 && <ChevronRightIcon key={index} width={16} />,
                curr.schema.title || curr.name,
              ],
              []
            )}
          </Row>
        }
      >
        <Typography.Text style={{ cursor: "help" }}>...</Typography.Text>
      </Tooltip>
      <ChevronRightIcon width={16} />
      <Typography.Text>{last.schema.title || last.name}</Typography.Text>
    </Row>
  ) : null;
}
