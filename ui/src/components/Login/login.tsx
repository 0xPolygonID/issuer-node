import { LockOutlined, UserOutlined } from "@ant-design/icons";
import { Button, Checkbox, Form, Input } from "antd";
// import "antd/dist/antd.css";
// import "./index.css";
import { generatePath, useNavigate } from "react-router-dom";
import { LoginLabel } from "src/domain";
import { ROUTES } from "src/routes";

export const Login = () => {
  const navigate = useNavigate();
  const onFinish = (values: LoginLabel) => {
    console.log("Received values of form: ", values);
    navigate(generatePath(ROUTES.connections.path));
  };

  return (
    <Form
      className="login-form"
      initialValues={{
        remember: true,
      }}
      name="normal_login"
      onFinish={onFinish}
    >
      <Form.Item
        name="username"
        rules={[
          {
            message: "Please input your Username!",
            required: true,
          },
        ]}
      >
        <Input placeholder="Username" prefix={<UserOutlined className="site-form-item-icon" />} />
      </Form.Item>
      <Form.Item
        name="password"
        rules={[
          {
            message: "Please input your Password!",
            required: true,
          },
        ]}
      >
        <Input
          placeholder="Password"
          prefix={<LockOutlined className="site-form-item-icon" />}
          type="password"
        />
      </Form.Item>
      <Form.Item>
        <Form.Item name="remember" noStyle valuePropName="checked">
          <Checkbox>Remember me</Checkbox>
        </Form.Item>

        <a className="login-form-forgot" href="">
          Forgot password
        </a>
      </Form.Item>

      <Form.Item>
        <Button className="login-form-button" htmlType="submit" type="primary">
          Log in
        </Button>
        Or <a href="">register now!</a>
      </Form.Item>
    </Form>
  );
};
