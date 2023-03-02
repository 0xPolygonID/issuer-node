import { Typography } from "antd";

import { COPYRIGHT_URL } from "src/utils/constants";

export function Copyright() {
  return (
    <Typography.Text type="secondary">
      {`© ${new Date().getFullYear()} `}
      <Typography.Link href={COPYRIGHT_URL}>Polygon ID</Typography.Link>
    </Typography.Text>
  );
}
