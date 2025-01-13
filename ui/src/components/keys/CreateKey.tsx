import { App, Button, Card, Divider, Flex, Form, Input, Select, Space } from "antd";
import { useNavigate } from "react-router-dom";

import { CreateKey as CreateKeyType, createKey } from "src/adapters/api/keys";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { KeyType } from "src/domain";
import { ROUTES } from "src/routes";
import { KEY_ADD_NEW, SAVE, VALUE_REQUIRED } from "src/utils/constants";

export function CreateKey() {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();
  const [form] = Form.useForm<CreateKeyType>();
  const navigate = useNavigate();
  const { message } = App.useApp();

  const handleSubmit = (formValues: CreateKeyType) => {
    return void createKey({
      env,
      identifier,
      payload: formValues,
    }).then((response) => {
      if (response.success) {
        void message.success("Key added successfully");
        navigate(ROUTES.keys.path);
      } else {
        void message.error(response.error.message);
      }
    });
  };

  return (
    <SiderLayoutContent
      description="Create a new display key"
      showBackButton
      showDivider
      title={KEY_ADD_NEW}
    >
      <Card className="centered" title="Key details">
        <Space direction="vertical" size="large">
          <Form
            form={form}
            initialValues={{
              keyType: KeyType.babyjubJub,
              name: "",
            }}
            layout="vertical"
            onFinish={handleSubmit}
          >
            <Form.Item
              label="Key name"
              name="name"
              rules={[{ message: VALUE_REQUIRED, required: true }]}
            >
              <Input placeholder="Enter name" />
            </Form.Item>

            <Form.Item
              label="Type"
              name="keyType"
              rules={[{ message: VALUE_REQUIRED, required: true }]}
            >
              <Select className="full-width" placeholder="Type">
                {Object.values(KeyType).map((type) => (
                  <Select.Option key={type} value={type}>
                    {type}
                  </Select.Option>
                ))}
              </Select>
            </Form.Item>

            <Divider />

            <Flex justify="flex-end">
              <Button htmlType="submit" type="primary">
                {SAVE}
              </Button>
            </Flex>
          </Form>
        </Space>
      </Card>
    </SiderLayoutContent>
  );
}
