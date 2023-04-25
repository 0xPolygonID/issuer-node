import { DownOutlined } from "@ant-design/icons";
import { Col, Tree } from "antd";
import type { DataNode } from "antd/es/tree";
import { useLayoutEffect, useRef, useState } from "react";

import { ObjectAttributeValueTreeNode } from "src/components/credentials/ObjectAttributeValueTreeNode";
import { AttributeValue, ObjectAttributeValue } from "src/domain";

const attributeToTreeDataNode = ({
  attributeValue,
  parents,
  treeWidth,
}: {
  attributeValue: AttributeValue;
  parents: ObjectAttributeValue[];
  treeWidth: number;
}): DataNode => {
  return {
    children:
      attributeValue.type === "object"
        ? attributeValue.value?.map((child) =>
            attributeToTreeDataNode({
              attributeValue: child,
              parents: [...parents, attributeValue],
              treeWidth,
            })
          )
        : [],
    key: [...parents, attributeValue].map((node) => node.name).join(" > "),
    title: (
      <ObjectAttributeValueTreeNode
        attributeValue={attributeValue}
        nestingLevel={parents.length}
        treeWidth={treeWidth}
      />
    ),
  };
};

export function ObjectAttributeValueTree({
  attributeValue,
  className,
  style,
}: {
  attributeValue: ObjectAttributeValue;
  className?: string;
  style?: React.CSSProperties;
}) {
  const [treeWidth, setTreeWidth] = useState(0);

  const ref = useRef<HTMLDivElement | null>(null);

  const treeData = [attributeToTreeDataNode({ attributeValue, parents: [], treeWidth })];

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
