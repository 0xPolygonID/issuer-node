import { Card } from "antd";
import { ReactNode } from "react";

export function Attribute({
  children,
  extra,
  index,
}: {
  children: ReactNode;
  extra?: ReactNode;
  index: number;
}) {
  return (
    <Card extra={extra} title={`Attribute #${index + 1}`} type="inner">
      {children}
    </Card>
  );
}
