import { Row, Typography } from "antd";

import { ReactComponent as IconCheckMark } from "src/assets/icons/check.svg";
import { ReactComponent as IconCopy } from "src/assets/icons/copy-01.svg";
import { DETAILS_MAXIMUM_WIDTH } from "src/utils/constants";

export function CopyableDetail({ data, label }: { data: string; label: string }) {
  return (
    <Row justify="space-between">
      <Typography.Text type="secondary">{label}</Typography.Text>

      <Typography.Text
        copyable={{ icon: [<IconCopy key={0} />, <IconCheckMark key={1} />] }}
        ellipsis
        style={{ maxWidth: DETAILS_MAXIMUM_WIDTH }}
      >
        {data}
      </Typography.Text>
    </Row>
  );
}
