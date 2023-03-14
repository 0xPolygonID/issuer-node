import {
  DatePicker,
  Form,
  FormItemProps,
  InputNumber,
  Radio,
  Select,
  Space,
  Typography,
} from "antd";

import { SchemaAttribute } from "src/adapters/api/schemas";
import { Attribute } from "src/components/schemas/Attribute";
import { DATE_VALIDITY_MESSAGE } from "src/utils/constants";

export function ClaimAttribute({
  index,
  schemaAttribute,
}: {
  index: number;
  schemaAttribute: SchemaAttribute;
}) {
  const { name, type } = schemaAttribute;
  const sharedFormItemProps: Partial<FormItemProps> = {
    label: (
      <>
        <Typography.Text ellipsis={{ tooltip: true }}>{name}</Typography.Text>
        {" *"}
      </>
    ),
    name: ["attributes", name],
  };

  return (
    <Attribute index={index}>
      {(() => {
        switch (type) {
          case "boolean": {
            return (
              <Form.Item
                {...sharedFormItemProps}
                rules={[{ message: "Value required", required: true }]}
              >
                <Radio.Group>
                  <Space direction="vertical">
                    <Radio value={0}>False (0)</Radio>
                    <Radio value={1}>True (1)</Radio>
                  </Space>
                </Radio.Group>
              </Form.Item>
            );
          }

          case "number": {
            return (
              <Form.Item
                {...sharedFormItemProps}
                rules={[
                  {
                    message: "Positive integer or 0 required",
                    required: true,
                  },
                ]}
              >
                <InputNumber className="full-width" min={0} type="number" />
              </Form.Item>
            );
          }

          case "date": {
            return (
              <Form.Item
                {...sharedFormItemProps}
                rules={[{ message: DATE_VALIDITY_MESSAGE, required: true }]}
              >
                <DatePicker />
              </Form.Item>
            );
          }

          case "singlechoice": {
            return (
              <Form.Item
                {...sharedFormItemProps}
                rules={[
                  {
                    message: "Value required",
                    required: true,
                  },
                ]}
              >
                <Select placeholder="Select option">
                  {schemaAttribute.values.map(({ name, value }, index) => (
                    <Select.Option key={index} value={value}>
                      {name} ({value})
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            );
          }
        }
      })()}
    </Attribute>
  );
}
