import { LockOutlined, UserOutlined } from "@ant-design/icons";
import { Button, Checkbox, Form, Input, message } from "antd";
import { generatePath, useNavigate } from "react-router-dom";
import { login } from "src/adapters/api/user";
import { useEnvContext } from "src/contexts/Env";
//import { useUserContext } from "src/contexts/UserDetails";
import { LoginLabel } from "src/domain";
import { ROUTES } from "src/routes";

export const Login = () => {
  const navigate = useNavigate();
  //const { setUserDetails } = useUserContext();
  const [messageAPI, messageContext] = message.useMessage();
  const env = useEnvContext();

  const onFinish = async (values: LoginLabel) => {
    console.log("Received values of form: ", values);
    if (values.username !== "issuer" && values.username !== "verifier") {
      try {
        const userDetails = await login({
          env,
          password: values.password,
          username: values.username,
        });

        if (userDetails.success) {
          localStorage.setItem("profile", userDetails.data.iscompleted.toString());
          navigate(generatePath(ROUTES.request.path));
          // setUserDetails(
          //   userDetails.data.username,
          //   userDetails.data.password,
          //   userDetails.data.userDID
          // );
        } else {
          void messageAPI.error("Wrong Credentials");
        }
      } catch (error) {
        // Handle the error, e.g., show an error message
        console.error("An error occurred:", error);
      }
    } else {
      localStorage.setItem("user", values.username);
      navigate(generatePath(ROUTES.request.path));
    }
  };

  return (
    <>
      {messageContext}
      <Form
        className="login-form"
        initialValues={{
          remember: true,
        }}
        name="normal_login"
        onFinish={() => onFinish}
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
      {/* {openProfileModal && <ProfileUpdateModal />} */}
      {/* onClose={() => setOpenModal(undefined)} */}
    </>
  );
};
