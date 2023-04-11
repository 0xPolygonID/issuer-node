import { Form, FormItemProps, Input } from "antd";

export function Text({ formItemProps }: { formItemProps: FormItemProps }): JSX.Element {
  return (
    <Form.Item {...formItemProps}>
      <Input />
    </Form.Item>
  );
}
