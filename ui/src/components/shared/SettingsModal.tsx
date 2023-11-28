import { Alert, Form, Input, Modal, Typography } from "antd";
import { useState } from "react";
import { z } from "zod";

import IconAlert from "src/assets/icons/alert-circle.svg?react";
import IconClose from "src/assets/icons/x.svg?react";
import { useEnvContext } from "src/contexts/Env";
import { theme } from "src/styles/theme";
import { getStorageByKey, setStorageByKey } from "src/utils/browser";
import {
  CLOSE,
  IPFS_CUSTOM_GATEWAY_KEY,
  IPFS_PUBLIC_GATEWAY_CHECKER_URL,
  SAVE,
  URL_FIELD_ERROR_MESSAGE,
} from "src/utils/constants";

export function SettingsModal({
  afterOpenChange,
  isOpen,
  onClose,
  onSave,
}: {
  afterOpenChange: (isOpen: boolean) => void;
  isOpen: boolean;
  onClose: () => void;
  onSave: () => void;
}) {
  const env = useEnvContext();
  const [form] = Form.useForm();

  const defaultIpfsGatewayUrl = getStorageByKey({
    defaultValue: env.ipfsGatewayUrl,
    key: IPFS_CUSTOM_GATEWAY_KEY,
    parser: z.string().url(),
  });

  const [areSettingsValid, setAreSettingsValid] = useState<boolean>(false);

  const onResetToDefault = () => {
    form.setFieldValue("ipfsGatewayUrl", env.ipfsGatewayUrl);
    void form.validateFields(["ipfsGatewayUrl"]);
  };

  const onSaveClick = () => {
    const parsedValue = z.string().url().safeParse(form.getFieldValue("ipfsGatewayUrl"));
    if (parsedValue.success) {
      const value = parsedValue.data.trim();
      const sanitizedValue = value.endsWith("/") ? value.slice(0, -1) : value;
      setStorageByKey({
        key: IPFS_CUSTOM_GATEWAY_KEY,
        value: sanitizedValue,
      });
      onSave();
    }
  };

  const onFormChange = () => {
    setAreSettingsValid(!form.getFieldsError().some(({ errors }) => errors.length));
  };

  return (
    <Modal
      afterOpenChange={afterOpenChange}
      cancelText={CLOSE}
      centered
      closable
      closeIcon={<IconClose />}
      maskClosable
      okButtonProps={{ disabled: !areSettingsValid }}
      okText={SAVE}
      onCancel={onClose}
      onOk={onSaveClick}
      open={isOpen}
      style={{ maxWidth: 400 }}
      title="Change IPFS gateway"
    >
      <Form
        form={form}
        initialValues={{
          ipfsGatewayUrl: defaultIpfsGatewayUrl,
        }}
        layout="vertical"
        onFieldsChange={onFormChange}
      >
        <Typography.Paragraph style={{ whiteSpace: "pre-line" }} type="secondary">
          The IPFS gateway makes it possible to access files hosted on IPFS. You can customize the
          IPFS gateway or continue using the default.{"\n"}
          {"\n"}You can use sites like the{" "}
          <Typography.Link href={IPFS_PUBLIC_GATEWAY_CHECKER_URL} target="_blank">
            IPFS Public Gateway Checker
          </Typography.Link>{" "}
          to choose a custom gateway.
        </Typography.Paragraph>
        <Form.Item
          extra={
            <Typography.Link
              onClick={onResetToDefault}
              style={{ display: "block", textAlign: "right" }}
            >
              Reset to default
            </Typography.Link>
          }
          label="IPFS gateway URL"
          name="ipfsGatewayUrl"
          required
          rules={[
            {
              required: true,
            },
            {
              message: URL_FIELD_ERROR_MESSAGE,
              validator: (_, value) => z.string().url().parseAsync(value),
            },
          ]}
        >
          <Input placeholder={env.ipfsGatewayUrl} />
        </Form.Item>
        <Alert
          icon={<IconAlert />}
          message="You might need to reload the page for the changes to be applied"
          showIcon
          style={{ background: "#FCFAFF", color: theme.token?.colorInfo, padding: 16 }}
          type="info"
        />
      </Form>
    </Modal>
  );
}
