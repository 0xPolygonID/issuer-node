import { Form, Radio, Select, Space, Typography } from "antd";

import { BooleanAttribute, ObjectAttribute } from "src/domain";

export function Boolean({
  attribute,
  parents,
}: {
  attribute: BooleanAttribute;
  parents: ObjectAttribute[];
}) {
  return (
    <Form.Item
      extra={attribute.schema.description}
      label={<Typography.Text>{attribute.schema.title || attribute.name}</Typography.Text>}
      name={["credentialSubject", ...parents.map((parent) => parent.name), attribute.name]}
      required={attribute.required}
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
