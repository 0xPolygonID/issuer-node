import {
  Card,
  DatePicker,
  Form,
  FormItemProps,
  Input,
  InputNumber,
  Radio,
  Select,
  Space,
  TimePicker,
  Typography,
} from "antd";
import { Fragment } from "react";

import { Attribute, ObjectAttribute } from "src/domain";
import { DATE_VALIDITY_MESSAGE } from "src/utils/constants";

function getSharedFormItemProps(
  attribute: Attribute,
  parents: Attribute[]
): Partial<FormItemProps> {
  const schema = attribute.type !== "multi" ? attribute.schema : attribute.schemas[0];

  return {
    label: (
      <Typography.Text ellipsis={{ tooltip: true }}>
        {(schema && schema.title) || attribute.name}
      </Typography.Text>
    ),
    name: ["attributes", ...parents.map((parent) => parent.name), attribute.name],
    required: attribute.required,
  };
}

export function CredentialForm({
  objectAttribute,
  parents,
}: {
  objectAttribute: ObjectAttribute;
  parents: Attribute[];
}) {
  function Enum({
    attribute,
    options,
  }: {
    attribute: Attribute;
    options: (string | number | boolean)[];
  }): JSX.Element {
    return (
      <Form.Item {...getSharedFormItemProps(attribute, parents)}>
        <Select placeholder="Select option">
          {options.map((option, index) => (
            <Select.Option key={index} value={option}>
              {option}
            </Select.Option>
          ))}
        </Select>
      </Form.Item>
    );
  }

  const isRoot = parents.length === 0;

  return (
    <Space direction="vertical" size="middle">
      {objectAttribute.schema.properties ? (
        objectAttribute.schema.properties.map((attribute) => {
          const key = [...parents, attribute].map((parent) => parent.name).join(" > ");
          const children = () => {
            switch (attribute.type) {
              case "boolean": {
                if (attribute.schema.enum) {
                  return <Enum attribute={attribute} options={attribute.schema.enum}></Enum>;
                } else {
                  return (
                    <Form.Item
                      {...getSharedFormItemProps(attribute, parents)}
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
              }

              case "number":
              case "integer": {
                if (attribute.schema.enum) {
                  return <Enum attribute={attribute} options={attribute.schema.enum}></Enum>;
                } else {
                  return (
                    <Form.Item
                      {...getSharedFormItemProps(attribute, parents)}
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
              }

              case "string": {
                if (attribute.schema.enum) {
                  return <Enum attribute={attribute} options={attribute.schema.enum}></Enum>;
                } else {
                  switch (attribute.schema.format) {
                    case "date":
                    case "date-time": {
                      return (
                        <Form.Item
                          {...getSharedFormItemProps(attribute, parents)}
                          rules={[{ message: DATE_VALIDITY_MESSAGE, required: true }]}
                        >
                          <DatePicker showTime={attribute.schema.format === "date-time"} />
                        </Form.Item>
                      );
                    }
                    case "time": {
                      return (
                        <Form.Item
                          {...getSharedFormItemProps(attribute, parents)}
                          rules={[{ message: DATE_VALIDITY_MESSAGE, required: true }]}
                        >
                          <TimePicker />
                        </Form.Item>
                      );
                    }
                    default: {
                      return (
                        <Form.Item {...getSharedFormItemProps(attribute, parents)}>
                          <Input />
                        </Form.Item>
                      );
                    }
                  }
                }
              }

              case "null": {
                return <Typography.Text>Null attributes are not yet supported</Typography.Text>;
              }

              case "array": {
                return <Typography.Text>Array attributes are not yet supported</Typography.Text>;
              }

              case "object": {
                return (
                  <Space direction="vertical">
                    <Typography.Text>{attribute.schema.title || attribute.name}</Typography.Text>
                    <CredentialForm objectAttribute={attribute} parents={[...parents, attribute]} />
                  </Space>
                );
              }

              case "multi": {
                return (
                  // ToDo: Implement multi-type schema attributes (PID-543)
                  <Typography.Text>Multi attributes are not yet supported</Typography.Text>
                );
              }
            }
          };
          return isRoot ? (
            <Card key={key}>{children()}</Card>
          ) : (
            <Fragment key={key}>{children()}</Fragment>
          );
        })
      ) : (
        <Typography.Text>The object has no properties defined</Typography.Text>
      )}
    </Space>
  );
}
