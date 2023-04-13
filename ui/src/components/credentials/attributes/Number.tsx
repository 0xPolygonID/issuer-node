import { Form, InputNumber, Select, Typography } from "antd";

import { IntegerAttribute, NumberAttribute, ObjectAttribute } from "src/domain";

export function Number({
  attribute,
  parents,
}: {
  attribute: IntegerAttribute | NumberAttribute;
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
        <InputNumber className="full-width" type="number" />
      )}
    </Form.Item>
  );
}
