import { Form, FormItemProps, Radio, Space } from "antd";

export function Boolean({ formItemProps }: { formItemProps: FormItemProps }): JSX.Element {
  return (
    <Form.Item {...formItemProps}>
      <Radio.Group>
        <Space direction="vertical">
          <Radio value={0}>False (0)</Radio>
          <Radio value={1}>True (1)</Radio>
        </Space>
      </Radio.Group>
    </Form.Item>
  );
}
