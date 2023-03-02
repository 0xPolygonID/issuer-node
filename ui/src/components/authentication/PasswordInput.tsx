import { Form, Input } from "antd";

import { ReactComponent as IconEyeClosed } from "src/assets/icons/eye-off.svg";
import { ReactComponent as IconEyeOpen } from "src/assets/icons/eye.svg";
import { ReactComponent as IconLock } from "src/assets/icons/lock-01.svg";

export function PasswordInput({ label }: { label?: string }) {
  return (
    <Form.Item
      label={`${label || "Password"}*`}
      name="password"
      rules={[
        {
          message: "Password required",
          required: true,
        },
      ]}
    >
      <Input.Password
        iconRender={(visible) => (visible ? <IconEyeClosed /> : <IconEyeOpen />)}
        placeholder="Enter your password"
        prefix={<IconLock className="icon-secondary" />}
      />
    </Form.Item>
  );
}
