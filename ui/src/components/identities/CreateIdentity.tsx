import { App, Card, Space } from "antd";
import { useNavigate } from "react-router-dom";
import { createIdentity } from "src/adapters/api/identities";
import { IdentityFormData } from "src/adapters/parsers/view";
import { IdentityForm } from "src/components/identities/IdentityForm";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { ROUTES } from "src/routes";
import { isAsyncTaskDataAvailable } from "src/utils/async";
import { IDENTITY_ADD, IDENTITY_ADD_NEW, IDENTITY_DETAILS } from "src/utils/constants";

export function CreateIdentity() {
  const env = useEnvContext();
  const { handleChange, identitiesList } = useIdentityContext();
  const { message } = App.useApp();
  const navigate = useNavigate();

  const handleSubmit = (formValues: IdentityFormData) => {
    const isUnique =
      isAsyncTaskDataAvailable(identitiesList) &&
      !identitiesList.data.some((identity) => identity.displayName === formValues.displayName);

    if (!isUnique) {
      return void message.error(`${formValues.displayName} is already exists`);
    }

    return void createIdentity({ env, payload: formValues }).then((response) => {
      if (response.success) {
        const {
          data: { identifier },
        } = response;

        void message.success("Identity added successfully");
        handleChange(identifier);
        navigate(ROUTES.identities.path);
      } else {
        void message.error(response.error.message);
      }
    });
  };

  return (
    <SiderLayoutContent
      description="View identity details and edit name"
      showBackButton
      showDivider
      title={IDENTITY_ADD_NEW}
    >
      <Card className="identities-card" title={IDENTITY_DETAILS}>
        <Space direction="vertical">
          <IdentityForm onSubmit={handleSubmit} submitBtnText={IDENTITY_ADD} />
        </Space>
      </Card>
    </SiderLayoutContent>
  );
}
