import { Divider } from "antd";

import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";

import { NOTIFICATION } from "src/utils/constants";

export function Notification() {
  return (
    <SiderLayoutContent title={NOTIFICATION}>
      <Divider />
    </SiderLayoutContent>
  );
}
