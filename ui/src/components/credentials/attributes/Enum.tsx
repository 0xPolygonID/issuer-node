import { Form, FormItemProps, Select } from "antd";

export function Enum({
  formItemProps,
  options,
}: {
  formItemProps: FormItemProps;
  options: (string | number | boolean)[];
}): JSX.Element {
  return (
    <Form.Item {...formItemProps}>
      <Select placeholder="Select option">
        {options.map((option, index) => (
          <Select.Option key={index} value={option}>
            {option}
          </Select.Option>
        ))}
      </Select>
    </Form.Item>
  );
}
