import { Col, Row, Space, Typography } from "antd";

import { ReactComponent as IconTreeLeaf } from "src/assets/icons/file-04.svg";
import { ReactComponent as IconTreeNode } from "src/assets/icons/folder.svg";
import { Schema } from "src/domain";

export function SchemaTreeNode({
  nestingLevel,
  schema,
  treeWidth,
}: {
  nestingLevel: number;
  schema: Schema;
  treeWidth: number;
}) {
  const commonProps = schema.type !== "multi" ? schema.schema : schema.schemas[0];
  const name = commonProps.title || schema.name;
  const description = commonProps.description;

  return (
    <Col>
      <Row justify="space-between">
        <Col>
          <Space>
            {schema.type === "object" ? (
              <IconTreeNode className="icon-secondary" />
            ) : (
              <IconTreeLeaf className="icon-secondary" />
            )}
            <Typography.Text>{name}</Typography.Text>

            {schema.required ? <Typography.Text type="danger">*</Typography.Text> : null}
          </Space>
        </Col>

        <Col style={{ marginLeft: 28 }}>
          {schema.type !== "object" && (
            <Typography.Text type="secondary">
              {schema.type === "string" && schema.schema.format
                ? schema.schema.format
                : schema.type}
            </Typography.Text>
          )}
        </Col>
      </Row>

      <Row>
        {description && (
          <Col style={{ marginLeft: 28 }}>
            <Typography.Text
              ellipsis={{ tooltip: true }}
              style={{
                maxWidth: treeWidth - 200 - nestingLevel * 20,
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
