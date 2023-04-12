import { Card, Space, Typography } from "antd";
import { Fragment, useRef } from "react";

import { AttributeBreadcrumb } from "./AttributeBreadcrumb";
import { Boolean } from "src/components/credentials/attributes/Boolean";
import { Number } from "src/components/credentials/attributes/Number";
import { String } from "src/components/credentials/attributes/String";
import { Attribute, ObjectAttribute } from "src/domain";

function AnyAttribute({
  attribute,
  parents,
}: {
  attribute: Attribute;
  parents: ObjectAttribute[];
}): JSX.Element {
  switch (attribute.type) {
    case "boolean": {
      return <Boolean attribute={attribute} parents={parents} />;
    }
    case "number":
    case "integer": {
      return <Number attribute={attribute} parents={parents} />;
    }
    case "string": {
      return <String attribute={attribute} parents={parents} />;
    }
    case "null": {
      return (
        <Typography.Text>Null attributes are not yet supported ({attribute.name})</Typography.Text>
      );
    }
    case "multi": {
      return (
        // ToDo: Implement multi-type schema attributes (PID-543)
        <Typography.Text>Multi attributes are not yet supported ({attribute.name})</Typography.Text>
      );
    }
    case "array": {
      return (
        <Typography.Text>Array attributes are not yet supported ({attribute.name})</Typography.Text>
      );
    }
    case "object": {
      return (
        <CredentialSubjectForm
          attributes={attribute.schema.properties || []}
          parents={[...parents, attribute]}
        />
      );
    }
  }
}

export function CredentialSubjectForm({
  attributes,
  parents = [],
}: {
  attributes: Attribute[];
  parents?: ObjectAttribute[];
}) {
  const shouldShowBreadcrumb = useRef<boolean>(true);
  return (
    <Space direction="vertical" size="middle">
      {[...attributes]
        .sort((a, b) =>
          a.type !== "object" && b.type !== "object" ? 0 : a.type === "object" ? 1 : -1
        )
        .map((attribute: Attribute) => {
          const showBreadcrumb = attribute.type !== "object" && shouldShowBreadcrumb.current;
          shouldShowBreadcrumb.current = !showBreadcrumb;
          const attributeNode = (
            <>
              {showBreadcrumb && <AttributeBreadcrumb parents={parents} />}
              <AnyAttribute attribute={attribute} parents={parents} />
            </>
          );
          const isRootAttribute = parents.length === 0;
          const shouldShowTitle = isRootAttribute && attribute.type === "object";
          const key = [...parents, attribute].map((parent) => parent.name).join(" > ");

          return isRootAttribute ? (
            <Card
              key={key}
              title={shouldShowTitle ? attribute.schema.title || attribute.name : undefined}
            >
              {attributeNode}
            </Card>
          ) : (
            <Fragment key={key}>{attributeNode}</Fragment>
          );
        })}
    </Space>
  );
}
