import { DownOutlined } from "@ant-design/icons";
import { Col, Tree } from "antd";
import type { DataNode } from "antd/es/tree";
import { useLayoutEffect, useRef, useState } from "react";

import { SchemaTreeNode } from "src/components/schemas/SchemaTreeNode";
import { Attribute, JsonSchema } from "src/domain/jsonSchema";

const attributeToTreeDataNode = ({
  attribute,
  parents,
  treeWidth,
}: {
  attribute: Attribute;
  parents: Attribute[];
  treeWidth: number;
}): DataNode => {
  return {
    children:
      attribute.type === "object" && attribute.schema.properties
        ? attribute.schema.properties.map((child) =>
            attributeToTreeDataNode({
              attribute: child,
              parents: [...parents, attribute],
              treeWidth,
            })
          )
        : [],
    key: [...parents, attribute].map((node) => node.name).join(" > "),
    title: (
      <SchemaTreeNode attribute={attribute} nestingLevel={parents.length} treeWidth={treeWidth} />
    ),
  };
};

export function SchemaTree({
  className,
  jsonSchema,
  style,
}: {
  className?: string;
  jsonSchema: JsonSchema;
  style?: React.CSSProperties;
}) {
  const [treeWidth, setTreeWidth] = useState(0);

  const ref = useRef<HTMLDivElement | null>(null);

  const treeData = [attributeToTreeDataNode({ attribute: jsonSchema, parents: [], treeWidth })];

  useLayoutEffect(() => {
    const updateWidth = () => {
      if (ref.current) {
        setTreeWidth(ref.current.offsetWidth);
      }
    };

    updateWidth();
    window.addEventListener("resize", updateWidth);

    return () => {
      window.removeEventListener("resize", updateWidth);
    };
  }, []);

  return (
    <Col ref={ref}>
      <Tree.DirectoryTree
        className={className}
        defaultExpandAll
        selectable={false}
        showIcon={false}
        showLine
        style={style}
        switcherIcon={<DownOutlined />}
        treeData={treeData}
      />
    </Col>
  );
}
