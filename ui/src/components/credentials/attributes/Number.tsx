import { Form, FormItemProps, InputNumber } from "antd";

export function Number({ formItemProps }: { formItemProps: FormItemProps }): JSX.Element {
  return (
    <Form.Item
      {...formItemProps}
      rules={[
        {
          message: "Positive integer or 0 required",
        },
      ]}
    >
      <InputNumber className="full-width" min={0} type="number" />
    </Form.Item>
  );
}
