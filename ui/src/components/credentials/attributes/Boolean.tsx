import { Form, Radio, Space, Typography } from "antd";

import { Enum } from "src/components/credentials/attributes/Enum";
import { BooleanAttribute, ObjectAttribute } from "src/domain";

export function Boolean({
  attribute,
  parents,
}: {
  attribute: BooleanAttribute;
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
      <Radio.Group>
        <Space direction="vertical">
          <Radio value={false}>False</Radio>
          <Radio value={true}>True</Radio>
        </Space>
      </Radio.Group>
    </Form.Item>
  );
}
