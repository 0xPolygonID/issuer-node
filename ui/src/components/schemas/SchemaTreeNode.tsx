import { Col, Row, Space, Typography } from "antd";

import IconTreeLeaf from "src/assets/icons/file-04.svg?react";
import IconTreeNode from "src/assets/icons/folder.svg?react";
import { Attribute } from "src/domain";

export function SchemaTreeNode({
  attribute,
  nestingLevel,
  treeWidth,
}: {
  attribute: Attribute;
  nestingLevel: number;
  treeWidth: number;
}) {
  const commonProps = attribute.type !== "multi" ? attribute.schema : attribute.schemas[0];
  const name = commonProps?.title || attribute.name;
  const description = commonProps?.description;

  return (
    <Col>
      <Row justify="space-between">
        <Col>
          <Space>
            {attribute.type === "object" ? (
              <IconTreeNode className="icon-secondary" />
            ) : (
              <IconTreeLeaf className="icon-secondary" />
            )}
            <Typography.Text>{name}</Typography.Text>

            {attribute.required ? <Typography.Text type="danger">*</Typography.Text> : null}
          </Space>
        </Col>

        <Col>
          {attribute.type !== "object" && (
            <Typography.Text type="secondary">
              {attribute.type === "string" && attribute.schema.format
                ? attribute.schema.format
                : attribute.type}
            </Typography.Text>
          )}
        </Col>
      </Row>

      <Row>
        {description && (
          <Col>
            <Typography.Text
              ellipsis={{ tooltip: true }}
              style={{
                maxWidth: treeWidth - name.length * 12 - nestingLevel * 20,
              }}
              type="secondary"
            >
              {description}
            </Typography.Text>
          </Col>
        )}
      </Row>
    </Col>
  );
}
