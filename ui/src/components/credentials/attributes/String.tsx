import { DatePicker, Form, FormItemProps, Input, Select, TimePicker, Typography } from "antd";

import { ObjectAttribute, StringAttribute } from "src/domain";
import { DATE_VALIDITY_MESSAGE, TIME_VALIDITY_MESSAGE } from "src/utils/constants";

export function String({
  attribute,
  parents,
}: {
  attribute: StringAttribute;
  parents: ObjectAttribute[];
}) {
  const commonFormItemProps: FormItemProps = {
    label: <Typography.Text>{attribute.schema.title || attribute.name}</Typography.Text>,
    name: ["attributes", ...parents.map((parent) => parent.name), attribute.name],
    required: attribute.required,
  };

  if (attribute.schema.enum) {
    return (
      <Form.Item {...commonFormItemProps}>
        <Select placeholder="Select option">
          {attribute.schema.enum.map((option, index) => (
            <Select.Option key={index} value={option}>
              {option}
            </Select.Option>
          ))}
        </Select>
      </Form.Item>
    );
  } else {
    switch (attribute.schema.format) {
      case "date":
      case "date-time": {
        return (
          <Form.Item
            {...commonFormItemProps}
            rules={[{ message: DATE_VALIDITY_MESSAGE, required: attribute.required }]}
          >
            <DatePicker showTime={attribute.schema.format === "date-time"} />
          </Form.Item>
        );
      }
      case "time": {
        return (
          <Form.Item
            {...commonFormItemProps}
            rules={[{ message: TIME_VALIDITY_MESSAGE, required: attribute.required }]}
          >
            <TimePicker />
          </Form.Item>
        );
      }
      default: {
        return (
          <Form.Item {...commonFormItemProps}>
            <Input />
          </Form.Item>
        );
      }
    }
  }
}
