import { Card, Space, Typography } from "antd";
import { Fragment, useRef } from "react";

import { AttributeBreadcrumb } from "src/components/credentials/AttributeBreadcrumb";
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
        // ToDo: PID-587
        <Typography.Text>Null attributes are not yet supported ({attribute.name})</Typography.Text>
      );
    }
    case "multi": {
      return (
        // ToDo: PID-543
        <Typography.Text>Multi attributes are not yet supported ({attribute.name})</Typography.Text>
      );
    }
    case "array": {
      return (
        // ToDo: PID-589
        <Typography.Text>Array attributes are not yet supported ({attribute.name})</Typography.Text>
      );
    }
    case "object": {
      return (
        <CredentialSubjectForm
          attributes={attribute.schema.properties || []}
          parents={[...parents, attribute]}
          space={false}
        />
      );
    }
  }
}

export function CredentialSubjectForm({
  attributes,
  parents = [],
  space = true,
}: {
  attributes: Attribute[];
  parents?: ObjectAttribute[];
  space?: boolean;
}) {
  const shouldShowBreadcrumb = useRef<boolean>(true);
  const form = [...attributes]
    .sort((a, b) => (a.type !== "object" && b.type !== "object" ? 0 : a.type === "object" ? 1 : -1))
    .map((attribute: Attribute) => {
      const showBreadcrumb =
        attribute.type !== "object" && parents.length > 1 && shouldShowBreadcrumb.current;

      shouldShowBreadcrumb.current = !showBreadcrumb;

      const attributeNode = showBreadcrumb ? (
        <Space direction="vertical" size="middle">
          <AttributeBreadcrumb parents={parents} />
          <AnyAttribute attribute={attribute} parents={parents} />
        </Space>
      ) : (
        <AnyAttribute attribute={attribute} parents={parents} />
      );

      const isRootAttribute = parents.length === 0;
      const shouldShowTitle = isRootAttribute && attribute.type === "object";
      const key = [...parents, attribute].map((parent) => parent.name).join(" > ");

      return isRootAttribute ? (
        <Card
          className="ant-card-type-inner"
          key={key}
          title={shouldShowTitle ? attribute.schema.title || attribute.name : undefined}
        >
          {attributeNode}
        </Card>
      ) : (
        <Fragment key={key}>{attributeNode}</Fragment>
      );
    });

  return space ? (
    <Space direction="vertical" size="large">
      {form}
    </Space>
  ) : (
    <>{form}</>
  );
}
