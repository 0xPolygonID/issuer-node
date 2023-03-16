import { Button, Divider, Space } from "antd";
import { generatePath, useNavigate } from "react-router-dom";

import { ReactComponent as IconCreditCardPlus } from "src/assets/icons/credit-card-plus.svg";
import { ReactComponent as IconUpload } from "src/assets/icons/upload-01.svg";
import { MySchemas } from "src/components/schemas/SchemasTable";
import { Explainer } from "src/components/shared/Explainer";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { ROUTES } from "src/routes";
import { IMPORT_SCHEMA, ISSUE_CREDENTIAL, SCHEMAS, TUTORIALS_URL } from "src/utils/constants";

export function Schemas() {
  const navigate = useNavigate();

  return (
    <SiderLayoutContent
      description="Verifiable credential schemas help to ensure the structure and data formatting across different services."
      extra={
        <Space align="start" size="middle">
          <Button
            icon={<IconCreditCardPlus />}
            onClick={() => navigate(generatePath(ROUTES.issueCredential.path))}
            type="default"
          >
            {ISSUE_CREDENTIAL}
          </Button>

          <Button
            icon={<IconUpload />}
            onClick={() => navigate(ROUTES.importSchema.path)}
            type="primary"
          >
            {IMPORT_SCHEMA}
          </Button>
        </Space>
      }
      title={SCHEMAS}
    >
      <Divider />

      <Space direction="vertical" size="large">
        <Explainer
          CTA={{ label: "Learn more", url: TUTORIALS_URL }}
          description="Learn about schema types, attributes, naming conventions, data types and more."
          image="/images/illustration-explainer.svg"
          title="Credential schemas explained"
        />

        <MySchemas />
      </Space>
    </SiderLayoutContent>
  );
}
