import { Card, Row, Space, Tooltip, Typography } from "antd";
import { Fragment, ReactNode } from "react";

import { ReactComponent as ChevronRightIcon } from "src/assets/icons/chevron-right.svg";
import { Boolean } from "src/components/credentials/attributes/Boolean";
import { Datetime } from "src/components/credentials/attributes/Datetime";
import { Enum } from "src/components/credentials/attributes/Enum";
import { Number } from "src/components/credentials/attributes/Number";
import { Text } from "src/components/credentials/attributes/Text";
import { Time } from "src/components/credentials/attributes/Time";
import { Attribute, AttributeSchema, ObjectAttribute } from "src/domain";

function getAttributeSchema(attribute: Attribute): AttributeSchema | undefined {
  return attribute.type === "multi" ? attribute.schemas[0] : attribute.schema;
}

function AnyAttribute({
  attribute,
  parents,
}: {
  attribute: Attribute;
  parents: ObjectAttribute[];
}): JSX.Element {
  const schema = getAttributeSchema(attribute);
  const formItemProps = {
    label: <Typography.Text>{(schema && schema.title) || attribute.name}</Typography.Text>,
    name: ["attributes", ...parents.map((parent) => parent.name), attribute.name],
    required: attribute.required,
  };

  switch (attribute.type) {
    case "boolean": {
      return attribute.schema.enum ? (
        <Enum formItemProps={formItemProps} options={attribute.schema.enum} />
      ) : (
        <Boolean formItemProps={formItemProps} />
      );
    }

    case "number":
    case "integer": {
      return attribute.schema.enum ? (
        <Enum formItemProps={formItemProps} options={attribute.schema.enum} />
      ) : (
        <Number formItemProps={formItemProps} />
      );
    }

    case "string": {
      if (attribute.schema.enum) {
        return <Enum formItemProps={formItemProps} options={attribute.schema.enum} />;
      } else {
        switch (attribute.schema.format) {
          case "date":
          case "date-time": {
            return (
              <Datetime
                formItemProps={formItemProps}
                showTime={attribute.schema.format === "date-time"}
              />
            );
          }
          case "time": {
            return <Time formItemProps={formItemProps} />;
          }
          default: {
            return <Text formItemProps={formItemProps} />;
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
        <CredentialForm
          attributes={attribute.schema.properties || []}
          breadcrumb={
            <Row align="bottom">
              <Tooltip
                placement="topLeft"
                title={
                  <Row align="bottom">
                    {parents.reduce(
                      (acc: React.ReactNode[], curr: ObjectAttribute, index): React.ReactNode[] => [
                        ...acc,
                        acc.length > 0 && <ChevronRightIcon key={index} width={16} />,
                        getAttributeSchema(curr)?.title || curr.name,
                      ],
                      []
                    )}
                  </Row>
                }
              >
                <Typography.Text style={{ cursor: "help" }}>...</Typography.Text>
              </Tooltip>
              <ChevronRightIcon width={16} />
              <Typography.Text>{attribute.schema.title || attribute.name}</Typography.Text>
            </Row>
          }
          parents={[...parents, attribute]}
        />
      );
    }

    case "multi": {
      return (
        // ToDo: Implement multi-type schema attributes (PID-543)
        <Typography.Text>Multi attributes are not yet supported</Typography.Text>
      );
    }
  }
}

export function CredentialForm({
  attributes,
  breadcrumb = null,
  parents = [],
}: {
  attributes: Attribute[];
  breadcrumb?: ReactNode;
  parents?: ObjectAttribute[];
}) {
  return (
    <Space direction="vertical" size="middle">
      {attributes.map((attribute: Attribute) => {
        const key = [...parents, attribute].map((parent) => parent.name).join(" > ");
        const form = (
          <>
            {attribute.type !== "object" && parents.length > 1 && breadcrumb}
            <AnyAttribute attribute={attribute} parents={parents} />
          </>
        );
        return parents.length === 0 ? (
          <Card
            key={key}
            title={
              attribute.type === "object" ? attribute.schema.title || attribute.name : undefined
            }
          >
            {form}
          </Card>
        ) : (
          <Fragment key={key}>{form}</Fragment>
        );
      })}
    </Space>
  );
}
