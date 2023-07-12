import { Form, Radio, Select, Space, Typography } from "antd";

import { BooleanAttribute, ObjectAttribute } from "src/domain";
import { VALUE_REQUIRED } from "src/utils/constants";

export function Boolean({
  attribute,
  error,
  parents,
}: {
  attribute: BooleanAttribute;
  error?: string;
  parents: ObjectAttribute[];
}) {
  return (
    <Form.Item
      extra={attribute.schema.description}
      help={error}
      label={<Typography.Text>{attribute.schema.title || attribute.name}</Typography.Text>}
      name={["credentialSubject", ...parents.map((parent) => parent.name), attribute.name]}
      rules={[{ message: VALUE_REQUIRED, required: attribute.required }]}
      validateStatus={error ? "error" : undefined}
    >
      {attribute.schema.enum ? (
        <Select placeholder="Select option">
          {attribute.schema.enum.map((option, index) => (
            <Select.Option key={index} value={option}>
              {option.toString()}
            </Select.Option>
          ))}
        </Select>
      ) : (
        <Radio.Group>
          <Space direction="vertical">
            <Radio value={true}>True</Radio>
            <Radio value={false}>False</Radio>
          </Space>
        </Radio.Group>
      )}
    </Form.Item>
  );
}
