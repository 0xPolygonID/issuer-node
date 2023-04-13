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
      label={<Typography.Text>{attribute.schema.title || attribute.name}</Typography.Text>}
      name={["attributes", ...parents.map((parent) => parent.name), attribute.name]}
      required={attribute.required}
    >
      {attribute.schema.enum ? (
        <Select placeholder="Select option">
          {attribute.schema.enum.map((option, index) => (
            <Select.Option key={index} value={option}>
              {option}
            </Select.Option>
          ))}
        </Select>
      ) : (
        <Radio.Group>
          <Space direction="vertical">
            <Radio value={false}>False</Radio>
            <Radio value={true}>True</Radio>
          </Space>
        </Radio.Group>
      )}
    </Form.Item>
  );
}
