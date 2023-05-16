import { Button, Col, Divider, Row, Space } from "antd";
import { generatePath, useNavigate } from "react-router-dom";

import { ReactComponent as IconCreditCardPlus } from "src/assets/icons/credit-card-plus.svg";
import { ReactComponent as IconUpload } from "src/assets/icons/upload-01.svg";
import { SchemasTable } from "src/components/schemas/SchemasTable";
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
        <Row gutter={[8, 8]}>
          <Col>
            <Button
              icon={<IconCreditCardPlus />}
              onClick={() => navigate(generatePath(ROUTES.issueCredential.path))}
              type="default"
            >
              {ISSUE_CREDENTIAL}
            </Button>
          </Col>

          <Col>
            <Button
              icon={<IconUpload />}
              onClick={() => navigate(ROUTES.importSchema.path)}
              type="primary"
            >
              {IMPORT_SCHEMA}
            </Button>
          </Col>
        </Row>
      }
      title={SCHEMAS}
    >
      <Divider />

      <Space direction="vertical" size="large">
        <Explainer
          CTA={{ label: "Learn more", url: TUTORIALS_URL }}
          description="Learn about schema types, attributes, naming conventions, data types and more."
          image="/images/illustration-explainer.svg"
          localStorageKey="explainerSchemas"
          title="Credential schemas explained"
        />

        <SchemasTable />
      </Space>
    </SiderLayoutContent>
  );
}
