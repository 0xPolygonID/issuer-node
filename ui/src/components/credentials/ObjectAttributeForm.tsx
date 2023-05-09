import { Card, Space, Typography } from "antd";
import { Fragment } from "react";

import { AttributeBreadcrumb } from "src/components/credentials/AttributeBreadcrumb";
import { Boolean } from "src/components/credentials/attributes/Boolean";
import { Number } from "src/components/credentials/attributes/Number";
import { String } from "src/components/credentials/attributes/String";
import { Attribute, ObjectAttribute } from "src/domain";

export type InputErrors = { [key: string]: string | InputErrors };

function AnyAttribute({
  attribute,
  inputErrors,
  parents,
}: {
  attribute: Attribute;
  inputErrors?: InputErrors;
  parents: ObjectAttribute[];
}): JSX.Element {
  const attributeError = inputErrors && inputErrors[attribute.name];
  const currentError = typeof attributeError === "string" ? attributeError : undefined;
  const nestedInputErrors = typeof attributeError !== "string" ? attributeError : undefined;
  switch (attribute.type) {
    case "boolean": {
      return <Boolean attribute={attribute} error={currentError} parents={parents} />;
    }
    case "number":
    case "integer": {
      return <Number attribute={attribute} error={currentError} parents={parents} />;
    }
    case "string": {
      return <String attribute={attribute} error={currentError} parents={parents} />;
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
        <ObjectAttributeForm
          attributes={attribute.schema.attributes || []}
          inputErrors={nestedInputErrors}
          parents={[...parents, attribute]}
        />
      );
    }
  }
}

export function ObjectAttributeForm({
  attributes,
  inputErrors,
  parents = [],
}: {
  attributes: Attribute[];
  inputErrors?: InputErrors;
  parents?: ObjectAttribute[];
}) {
  const isRootAttribute = parents.length === 0;
  const form = attributes.map((attribute: Attribute, index) => {
    const showBreadcrumb = attribute.type !== "object" && parents.length > 1 && index === 0;

    const attributeNode = showBreadcrumb ? (
      <Space direction="vertical" size="middle">
        <AttributeBreadcrumb parents={parents} />

        <AnyAttribute attribute={attribute} inputErrors={inputErrors} parents={parents} />
      </Space>
    ) : (
      <AnyAttribute attribute={attribute} inputErrors={inputErrors} parents={parents} />
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
