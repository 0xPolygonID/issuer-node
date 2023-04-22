import { Col, Row, Space, Typography } from "antd";

import { z } from "zod";
import { ReactComponent as IconTreeLeaf } from "src/assets/icons/file-04.svg";
import { ReactComponent as IconTreeNode } from "src/assets/icons/folder.svg";
import { AttributeValue, JsonLiteral } from "src/domain";
import { formatDate } from "src/utils/forms";

function extractValue(attributeValue: AttributeValue): JsonLiteral | undefined {
  switch (attributeValue.type) {
    case "integer":
    case "number":
    case "null":
    case "boolean": {
      return attributeValue.value === undefined ? "-" : attributeValue.value;
    }
    case "string": {
      const value = attributeValue.value;
      if (value === undefined) {
        return "-";
      } else {
        switch (attributeValue.schema.format) {
          case "date":
          case "date-time":
          case "time": {
            const parsedDate = z.coerce.date(z.string().datetime()).safeParse(value);
            return parsedDate.success
              ? formatDate(parsedDate.data, attributeValue.schema.format)
              : attributeValue.value;
          }
          default: {
            return attributeValue.value === undefined ? "-" : attributeValue.value;
          }
        }
      }
    }
    case "object": {
      return undefined;
    }
    case "array": {
      return attributeValue.value?.map(extractValue).join(", ") || "-";
    }
    case "multi": {
      return attributeValue.value === undefined ? "-" : extractValue(attributeValue.value);
    }
  }
}

export function ObjectAttributeValuesTreeNode({
  attributeValue,
  nestingLevel,
  treeWidth,
}: {
  attributeValue: AttributeValue;
  nestingLevel: number;
  treeWidth: number;
}) {
  const commonProps =
    attributeValue.type !== "multi" ? attributeValue.schema : attributeValue.schemas[0];
  const name = commonProps?.title || attributeValue.name;

  return (
    <Col>
      <Row justify="space-between">
        <Col>
          <Space>
            {attributeValue.type === "object" ? (
              <IconTreeNode className="icon-secondary" />
            ) : (
              <IconTreeLeaf className="icon-secondary" />
            )}
            <Typography.Text>{name}</Typography.Text>
          </Space>
        </Col>

        <Col style={{ marginLeft: 28 }}>
          <Typography.Text
            ellipsis={{ tooltip: true }}
            style={{
              maxWidth: treeWidth - 200 - nestingLevel * 20,
            }}
            type="secondary"
          >
            {extractValue(attributeValue)}
          </Typography.Text>
        </Col>
      </Row>
    </Col>
  );
}
