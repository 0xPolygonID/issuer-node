import { Button, Form, Space, Typography } from "antd";
import { useState } from "react";
import { useNavigate } from "react-router-dom";

import { Authentication, authenticationSignIn, parseAccount } from "src/adapters/api/accounts";
import { EmailInput } from "src/components/authentication/EmailInput";
import { PasswordInput } from "src/components/authentication/PasswordInput";
import { ErrorMessage } from "src/components/shared/ErrorMessage";
import { useAuthContext } from "src/hooks/useAuthContext";
import { APIError } from "src/utils/adapters";
import { ROOT_PATH } from "src/utils/constants";
import { AsyncTask } from "src/utils/types";

export function SignIn() {
  const { updateAuthToken } = useAuthContext();

  const navigate = useNavigate();
  const [signIn, setSignIn] = useState<AsyncTask<never, APIError>>({
    status: "pending",
  });

  const onValidSubmit = (payload: Authentication) => {
    setSignIn({ status: "loading" });
    void authenticationSignIn({ payload }).then((response) => {
      if (response.isSuccessful) {
        const { token } = response.data;
        const { organization, verified } = parseAccount(token);

        if (verified) {
          updateAuthToken(token);

          if (organization) {
            navigate(ROOT_PATH);
          }
        }
      } else {
        setSignIn({ error: response.error, status: "failed" });
      }
    });
  };

  return (
    <Space direction="vertical" size="large">
      <Typography.Title level={2}>Sign in to your Polygon ID account</Typography.Title>

      <Form
        layout="vertical"
        name="sign-in"
        onFinish={onValidSubmit}
        requiredMark={false}
        validateTrigger="onBlur"
      >
        {signIn.status === "failed" && <ErrorMessage text={signIn.error.message} />}

        <EmailInput />

        <PasswordInput />

        <Form.Item>
          <Button block htmlType="submit" loading={signIn.status === "loading"} type="primary">
            Sign in
          </Button>
        </Form.Item>
      </Form>
    </Space>
  );
}
