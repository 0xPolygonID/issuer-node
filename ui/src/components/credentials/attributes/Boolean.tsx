import { Form, Radio, Select, Space, Typography } from "antd";

import IconInfoCircle from "src/assets/icons/info-circle.svg?react";
import { BooleanAttribute, ObjectAttribute } from "src/domain";
import { VALUE_REQUIRED } from "src/utils/constants";

export function Boolean({
  attribute,
  disabled = false,
  error,
  parents,
}: {
  attribute: BooleanAttribute;
  disabled?: boolean;
  error?: string;
  parents: ObjectAttribute[];
}) {
  return (
    <Form.Item
      help={error}
      label={<Typography.Text>{attribute.schema.title || attribute.name}</Typography.Text>}
      name={["credentialSubject", ...parents.map((parent) => parent.name), attribute.name]}
      rules={[{ message: VALUE_REQUIRED, required: attribute.required }]}
      tooltip={{
        icon: <IconInfoCircle style={{ width: 14 }} />,
        placement: "right",
        title: attribute.schema.description,
      }}
      validateStatus={error ? "error" : undefined}
    >
      {attribute.schema.enum ? (
        <Select disabled={disabled} placeholder="Select option">
          {attribute.schema.enum.map((option, index) => (
            <Select.Option key={index} value={option}>
              {option.toString()}
            </Select.Option>
          ))}
        </Select>
      ) : (
        <Radio.Group disabled={disabled}>
          <Space direction="vertical">
            <Radio value={true}>True</Radio>
            <Radio value={false}>False</Radio>
          </Space>
        </Radio.Group>
      )}
    </Form.Item>
  );
}
