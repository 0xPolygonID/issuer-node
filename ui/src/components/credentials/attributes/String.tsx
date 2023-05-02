import { DatePicker, Form, FormItemProps, Input, Select, TimePicker, Typography } from "antd";

import { ObjectAttribute, StringAttribute } from "src/domain";
import { VALUE_REQUIRED } from "src/utils/constants";

export function String({
  attribute,
  parents,
}: {
  attribute: StringAttribute;
  parents: ObjectAttribute[];
}) {
  const commonFormItemProps: FormItemProps = {
    extra: attribute.schema.description,
    label: <Typography.Text>{attribute.schema.title || attribute.name}</Typography.Text>,
    name: ["credentialSubject", ...parents.map((parent) => parent.name), attribute.name],
    rules: [{ message: VALUE_REQUIRED, required: attribute.required }],
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
          <Form.Item {...commonFormItemProps}>
            <DatePicker showTime={attribute.schema.format === "date-time"} />
          </Form.Item>
        );
      }
      case "time": {
        return (
          <Form.Item {...commonFormItemProps}>
            <TimePicker />
          </Form.Item>
        );
      }
      default: {
        return (
          <Form.Item {...commonFormItemProps}>
            <Input placeholder={`Type ${attribute.schema.format || attribute.type}`} />
          </Form.Item>
        );
      }
    }
  }
}
