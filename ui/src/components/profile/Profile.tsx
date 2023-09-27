import { Divider } from "antd";

import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";

import { PROFILE } from "src/utils/constants";

export function Profile() {
  return (
    <SiderLayoutContent title={PROFILE}>
      <Divider />
    </SiderLayoutContent>
  );
}
