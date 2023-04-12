import { Form, InputNumber, Typography } from "antd";

import { Enum } from "src/components/credentials/attributes/Enum";
import { IntegerAttribute, NumberAttribute, ObjectAttribute } from "src/domain";

export function Number({
  attribute,
  parents,
}: {
  attribute: IntegerAttribute | NumberAttribute;
  parents: ObjectAttribute[];
}) {
  return attribute.schema.enum ? (
    <Enum attribute={attribute} parents={parents} />
  ) : (
    <Form.Item
      label={<Typography.Text>{attribute.schema.title || attribute.name}</Typography.Text>}
      name={["attributes", ...parents.map((parent) => parent.name), attribute.name]}
      required={attribute.required}
    >
      <InputNumber className="full-width" type="number" />
    </Form.Item>
  );
}
