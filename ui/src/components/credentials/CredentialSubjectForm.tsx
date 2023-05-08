import { Card, Space, Typography } from "antd";
import { Fragment } from "react";

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
  const isRootAttribute = parents.length === 0;
  const form = attributes.map((attribute: Attribute, index) => {
    const showBreadcrumb = attribute.type !== "object" && parents.length > 1 && index === 0;

    const attributeNode = showBreadcrumb ? (
      <Space direction="vertical" size="middle">
        <AttributeBreadcrumb parents={parents} />

        <AnyAttribute attribute={attribute} parents={parents} />
      </Space>
    ) : (
      <AnyAttribute attribute={attribute} parents={parents} />
    );

    const shouldShowTitle = isRootAttribute && attribute.type === "object";
    const key = [...parents, attribute].map((parent) => parent.name).join(" > ");

    return isRootAttribute ? (
      <Card
        key={key}
        title={shouldShowTitle ? attribute.schema.title || attribute.name : undefined}
        type="inner"
      >
        {attributeNode}
      </Card>
    ) : (
      <Fragment key={key}>{attributeNode}</Fragment>
    );
  });

  return isRootAttribute ? (
    <Space direction="vertical" size="large">
      {form}
    </Space>
  ) : (
    <>{form}</>
  );
}
