import { Form, Select, Typography } from "antd";

import {
  BooleanAttribute,
  IntegerAttribute,
  NumberAttribute,
  ObjectAttribute,
  StringAttribute,
} from "src/domain";

export function Enum({
  attribute,
  parents,
}: {
  attribute: StringAttribute | NumberAttribute | IntegerAttribute | BooleanAttribute;
  parents: ObjectAttribute[];
}) {
  return attribute.schema.enum ? (
    <Form.Item
      label={<Typography.Text>{attribute.schema.title || attribute.name}</Typography.Text>}
      name={["attributes", ...parents.map((parent) => parent.name), attribute.name]}
      required={attribute.required}
    >
      <Select placeholder="Select option">
        {attribute.schema.enum.map((option, index) => (
          <Select.Option key={index} value={option}>
            {option}
          </Select.Option>
        ))}
      </Select>{" "}
    </Form.Item>
  ) : (
    <Typography.Text>No options are available</Typography.Text>
  );
}
