import { Typography } from "antd";

export function ErrorMessage({ text }: { text: string }) {
  return (
    <Typography.Title level={5} style={{ marginBottom: 16 }} type="danger">
      {text}
    </Typography.Title>
  );
}
