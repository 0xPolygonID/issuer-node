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
            const parsedDate = z.coerce.date().safeParse(value);
            return parsedDate.success
              ? formatDate(parsedDate.data, attributeValue.schema.format)
              : value;
          }
          default: {
            return value;
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

export function ObjectAttributeValueTreeNode({
  attributeValue,
  ellipsisPosition,
  nestingLevel,
  treeWidth,
}: {
  attributeValue: AttributeValue;
  ellipsisPosition?: number;
  nestingLevel: number;
  treeWidth: number;
}) {
  const commonProps =
    attributeValue.type !== "multi" ? attributeValue.schema : attributeValue.schemas[0];
  const name = commonProps?.title || attributeValue.name;
  const rawValue = extractValue(attributeValue);
  const stringValue =
    rawValue === null ? "null" : rawValue === undefined ? undefined : rawValue.toString();
  const slicedValue =
    stringValue !== undefined && ellipsisPosition
      ? stringValue.slice(0, stringValue.length - ellipsisPosition)
      : stringValue;

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

        <Col>
          <Typography.Text
            ellipsis={
              stringValue && ellipsisPosition
                ? { suffix: stringValue.slice(-ellipsisPosition), tooltip: { title: stringValue } }
                : { tooltip: { title: stringValue } }
            }
            style={{
              maxWidth:
                slicedValue &&
                treeWidth - name.length * 12 - nestingLevel * 24 >= slicedValue.length * 3
                  ? treeWidth - name.length * 12 - nestingLevel * 24
                  : treeWidth - nestingLevel * 24 - 32,
            }}
            type="secondary"
          >
            {slicedValue}
          </Typography.Text>
        </Col>
      </Row>
    </Col>
  );
}
