import { Button, Col, Divider, Row, Space, Typography } from "antd";
import { generatePath, useNavigate } from "react-router-dom";

import IconCreditCardPlus from "src/assets/icons/credit-card-plus.svg?react";
import IconUpload from "src/assets/icons/upload-01.svg?react";
import { SchemasTable } from "src/components/schemas/SchemasTable";
import { Explainer } from "src/components/shared/Explainer";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { ROUTES } from "src/routes";
import { IMPORT_SCHEMA, ISSUE_CREDENTIAL, SCHEMAS, SCHEMAS_DOCS_URL } from "src/utils/constants";

export function Schemas() {
  const navigate = useNavigate();
  const env = useEnvContext();
  return (
    <SiderLayoutContent
      description={
        <>
          <Typography.Text type="secondary">
            Verifiable credential schemas help to ensure the structure and data formatting across
            different services.
          </Typography.Text>
          {env.schemaExplorerAndBuilderUrl && (
            <>
              {" "}
              <Typography.Text type="secondary">
                Explore a wide range of existing schemas or create custom schemas using
              </Typography.Text>{" "}
              <Typography.Link href={env.schemaExplorerAndBuilderUrl} target="_blank">
                Privado ID&apos;s schema explorer and builder.
              </Typography.Link>
            </>
          )}
        </>
      }
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
          CTA={{ label: "Learn more", url: SCHEMAS_DOCS_URL }}
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
