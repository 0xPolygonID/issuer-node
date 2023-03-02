import { LoadingOutlined } from "@ant-design/icons";
import { Spin } from "antd";
import { SpinSize } from "antd/lib/spin";

export function Spinner({ size = "large" }: { size?: SpinSize }) {
  return <Spin indicator={<LoadingOutlined spin />} size={size} />;
}
