import { App, Card, Space } from "antd";
import { useNavigate } from "react-router-dom";

import { UpsertDisplayMethod, createDisplayMethod } from "src/adapters/api/display-method";
import { DisplayMethodForm } from "src/components/display-methods/DisplayMethodForm";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { ROUTES } from "src/routes";
import { DISPLAY_METHOD_ADD_NEW } from "src/utils/constants";

export function CreateDisplayMethod() {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();
  const navigate = useNavigate();
  const { message } = App.useApp();

  const handleSubmit = (formValues: UpsertDisplayMethod) => {
    return void createDisplayMethod({
      env,
      identifier,
      payload: { ...formValues, name: formValues.name.trim() },
    }).then((response) => {
      if (response.success) {
        void message.success("Display method added successfully");
        navigate(ROUTES.displayMethods.path);
      } else {
        void message.error(response.error.message);
      }
    });
  };

  return (
    <SiderLayoutContent
      description="Create and customize a new display method"
      showBackButton
      showDivider
      title={DISPLAY_METHOD_ADD_NEW}
    >
      <Card className="centered" title="Display method details">
        <Space direction="vertical" size="large">
          <DisplayMethodForm
            initialValues={{
              name: "",
              url: "",
            }}
            onSubmit={handleSubmit}
          />
        </Space>
      </Card>
    </SiderLayoutContent>
  );
}
