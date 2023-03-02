import { Form, Input } from "antd";

import { ReactComponent as IconMail } from "src/assets/icons/mail-01.svg";
import { emailValidator } from "src/utils/forms";

export function EmailInput({ label }: { label?: string }) {
  return (
    <Form.Item
      label={`${label || "Email address"}*`}
      name="email"
      rules={[
        {
          message: "Valid email address required",
          required: true,
          validator: emailValidator,
        },
      ]}
    >
      <Input placeholder="example@email.com" prefix={<IconMail className="icon-secondary" />} />
    </Form.Item>
  );
}
