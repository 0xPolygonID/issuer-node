import { Card, Space, message } from "antd";
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
  const { identityList, selectIdentity } = useIdentityContext();
  const [messageAPI, messageContext] = message.useMessage();
  const navigate = useNavigate();

  const handleSubmit = (formValues: IdentityFormData) => {
    const isUnique =
      isAsyncTaskDataAvailable(identityList) &&
      !identityList.data.some((identity) => identity.displayName === formValues.displayName);

    if (!isUnique) {
      return void messageAPI.error(`${formValues.displayName} already exists`);
    }

    return void createIdentity({
      env,
      payload: { ...formValues, displayName: formValues.displayName.trim() },
    }).then((response) => {
      if (response.success) {
        const {
          data: { identifier },
        } = response;

        void messageAPI.success("Identity added successfully");
        selectIdentity(identifier);
        navigate(ROUTES.identities.path);
      } else {
        void messageAPI.error(response.error.message);
      }
    });
  };

  return (
    <>
      {messageContext}

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
    </>
  );
}
