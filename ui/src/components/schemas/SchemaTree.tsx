import { DownOutlined } from "@ant-design/icons";
import { Col, Tree } from "antd";
import type { DataNode } from "antd/es/tree";
import { useLayoutEffect, useRef, useState } from "react";

import { SchemaTreeNode } from "src/components/schemas/SchemaTreeNode";
import { Schema } from "src/domain";

const schemaToTreeDataNodes = ({
  parents,
  schema,
  treeWidth,
}: {
  parents: Schema[];
  schema: Schema;
  treeWidth: number;
}): DataNode[] => {
  const children = (schema.type === "object" && schema.schema.properties) || [];

  return children.map((child) => ({
    children:
      child.type === "object"
        ? schemaToTreeDataNodes({ parents: [...parents, schema], schema: child, treeWidth })
        : [],
    key: [...parents, schema, child].map((node) => node.name).join(" > "),
    title: <SchemaTreeNode nestingLevel={parents.length} schema={child} treeWidth={treeWidth} />,
  }));
};

export function SchemaTree({ schema, style }: { schema: Schema; style?: React.CSSProperties }) {
  const ref = useRef<HTMLDivElement | null>(null);
  const [treeWidth, setTreeWidth] = useState(0);

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

  const treeData = [
    {
      children: schemaToTreeDataNodes({ parents: [], schema, treeWidth }),
      key: schema.name,
      title: <SchemaTreeNode nestingLevel={0} schema={schema} treeWidth={treeWidth} />,
    },
  ];

  return (
    <Col ref={ref}>
      <Tree.DirectoryTree
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
