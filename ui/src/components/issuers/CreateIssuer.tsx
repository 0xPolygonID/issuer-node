import { Card, Space, message } from "antd";
import { useNavigate } from "react-router-dom";
import { createIssuer } from "src/adapters/api/issuers";
import { IssuerFormData } from "src/adapters/parsers/view";
import { IssuerForm } from "src/components/issuers/IssuerForm";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { ROUTES } from "src/routes";
import { ISSUER_ADD, ISSUER_DETAILS } from "src/utils/constants";

export function CreateIssuer() {
  const env = useEnvContext();
  const [messageAPI, messageContext] = message.useMessage();
  const navigate = useNavigate();

  const handleSubmit = (formValues: IssuerFormData) =>
    void createIssuer({ env, payload: formValues }).then((response) => {
      if (response.success) {
        navigate(ROUTES.issuers.path);

        void messageAPI.success("Issuer added");
      } else {
        void messageAPI.error(response.error.message);
      }
    });

  return (
    <>
      {messageContext}

      <SiderLayoutContent
        description="Add a new issuer to get the required credential."
        showBackButton
        showDivider
        title={ISSUER_ADD}
      >
        <Card className="issuers-card" title={ISSUER_DETAILS}>
          <Space direction="vertical">
            <IssuerForm onBack={() => navigate(ROUTES.issuers.path)} onSubmit={handleSubmit} />
          </Space>
        </Card>
      </SiderLayoutContent>
    </>
  );
}
