import { Button, Col, Form, Input, InputNumber, Row, Select, Space } from "antd";
import { useState } from "react";

import { SchemaAttribute } from "src/adapters/api/schemas";
import { ReactComponent as IconAdd } from "src/assets/icons/plus.svg";
import { ReactComponent as IconRemove } from "src/assets/icons/trash-01.svg";
import { Attribute } from "src/components/schemas/Attribute";
import {
  SCHEMA_FORM_EXTRA_ALPHA_MESSAGE,
  SCHEMA_FORM_HELP_ALPHA_MESSAGE,
  SCHEMA_FORM_HELP_NUMERIC_MESSAGE,
  SCHEMA_FORM_HELP_REQUIRED_MESSAGE,
  SCHEMA_KEY_MAX_LENGTH,
} from "src/utils/constants";
import { alphanumericValidator } from "src/utils/forms";

const MESSAGE_HELP_DATA_TYPE: Record<SchemaAttribute["type"], string> = {
  boolean: "Choose true or false when issuing a claim.",
  date: "Input a date when issuing a claim.",
  number: "Input a positive integer or 0 when issuing a claim.",
  singlechoice: "Choose a single option from listed values when issuing a claim.",
};

const TYPES_DATA: { label: string; value: SchemaAttribute["type"] }[] = [
  { label: "True / False (Boolean)", value: "boolean" },
  { label: "Date", value: "date" },
  { label: "Number", value: "number" },
  { label: "Single Choice", value: "singlechoice" },
];

export function SchemaAttributeField({
  index,
  onAttributeTypeChange,
  onRemove,
  singleChoiceValuesName,
}: {
  index: number;
  onAttributeTypeChange: (index: number) => void;
  onRemove?: (index: number | number[]) => void;
  singleChoiceValuesName: string;
}) {
  const [attributeType, setAttributeType] = useState<SchemaAttribute["type"]>("boolean");

  const onSelectChange = (dataType: SchemaAttribute["type"]) => {
    setAttributeType(dataType);
    onAttributeTypeChange(index);
  };

  return (
    <Attribute
      extra={
        onRemove && (
          <Button
            danger
            icon={<IconRemove />}
            onClick={() => onRemove(index)}
            size="small"
            style={{ marginLeft: "auto" }}
            type="link"
          >
            Remove
          </Button>
        )
      }
      index={index}
    >
      <Form.Item
        extra="User-friendly name that describes the schema. Shown to end-users."
        label="Display name*"
        name={[index, "name"]}
        rules={[{ message: SCHEMA_FORM_HELP_REQUIRED_MESSAGE, required: true }]}
      >
        <Input maxLength={SCHEMA_KEY_MAX_LENGTH} placeholder="e.g. Attribute Name" showCount />
      </Form.Item>

      <Form.Item
        extra={SCHEMA_FORM_EXTRA_ALPHA_MESSAGE}
        label="Hidden name*"
        name={[index, "technicalName"]}
        rules={[
          {
            message: SCHEMA_FORM_HELP_REQUIRED_MESSAGE,
            required: true,
          },
          {
            message: SCHEMA_FORM_HELP_ALPHA_MESSAGE,
            validator: alphanumericValidator,
          },
        ]}
      >
        <Input maxLength={SCHEMA_KEY_MAX_LENGTH} placeholder="e.g. attributeName" showCount />
      </Form.Item>

      <Form.Item
        extra={MESSAGE_HELP_DATA_TYPE[attributeType]}
        label="Data type*"
        name={[index, "type"]}
        rules={[{ message: "Data type selection is required.", required: true }]}
      >
        <Select onChange={onSelectChange}>
          {TYPES_DATA.map(({ label, value }, index) => (
            <Select.Option key={index} value={value}>
              {label}
            </Select.Option>
          ))}
        </Select>
      </Form.Item>

      {attributeType === "singlechoice" && (
        <Form.List
          initialValue={[
            {
              name: "",
              value: "",
            },
          ]}
          name={[index, singleChoiceValuesName]}
        >
          {(fields, { add, remove }) => (
            <Space direction="vertical">
              {fields.map(({ key, name }, i) => (
                <Row align="top" gutter={16} key={key}>
                  <Col span={11}>
                    <Form.Item
                      extra={fields.length === i + 1 && SCHEMA_FORM_HELP_ALPHA_MESSAGE}
                      label={i === 0 && "Value*"}
                      name={[name, "name"]}
                      rules={[
                        {
                          message: SCHEMA_FORM_HELP_REQUIRED_MESSAGE,
                          required: true,
                        },
                        {
                          message: SCHEMA_FORM_HELP_ALPHA_MESSAGE,
                          validator: alphanumericValidator,
                        },
                      ]}
                    >
                      <Input maxLength={SCHEMA_KEY_MAX_LENGTH} placeholder="Type text" />
                    </Form.Item>
                  </Col>

                  <Col span={11}>
                    <Form.Item
                      extra={fields.length === i + 1 && SCHEMA_FORM_HELP_NUMERIC_MESSAGE}
                      label={i === 0 && "Numeric value*"}
                      name={[name, "value"]}
                      rules={[{ message: SCHEMA_FORM_HELP_REQUIRED_MESSAGE, required: true }]}
                    >
                      <InputNumber
                        className="full-width"
                        maxLength={SCHEMA_KEY_MAX_LENGTH}
                        min={0}
                        placeholder="e.g. 0"
                        precision={0}
                        size="large"
                        type="number"
                      />
                    </Form.Item>
                  </Col>

                  <Col span={2}>
                    {fields.length > 1 && (
                      <Form.Item label={i === 0 && " "}>
                        <Button
                          icon={<IconRemove className="icon-secondary" />}
                          onClick={() => remove(key)}
                          type="text"
                        />
                      </Form.Item>
                    )}
                  </Col>
                </Row>
              ))}

              <Button icon={<IconAdd />} onClick={add} size="small" type="link">
                Add value
              </Button>
            </Space>
          )}
        </Form.List>
      )}

      <Form.Item label="Description (Optional)" name={[index, "description"]}>
        <Input.TextArea placeholder="Enter a description..." rows={4} />
      </Form.Item>
    </Attribute>
  );
}
