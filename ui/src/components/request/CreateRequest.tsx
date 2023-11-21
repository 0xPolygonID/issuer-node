import { Button, Card, Form, Input, Select } from "antd";
import { useEffect, useState } from "react";
import { generatePath, useNavigate } from "react-router-dom";
import { requestVC } from "src/adapters/api/requests";
import { getAllSchema } from "src/adapters/api/schemas";
import { getUser } from "src/adapters/api/user";

import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { useUserContext } from "src/contexts/UserDetails";
import { AppError } from "src/domain";
import { Schema } from "src/domain/schema";
import { FormData, UserDetails } from "src/domain/user";
import { ROUTES } from "src/routes";
import { AsyncTask, isAsyncTaskDataAvailable } from "src/utils/async";
import { isAbortedError } from "src/utils/browser";

import { CREATE_REQUEST, SCHEMA_TYPE, SUBMIT, VALUE_REQUIRED } from "src/utils/constants";
import { notifyParseErrors } from "src/utils/error";

export function CreateRequest() {
  const navigate = useNavigate();
  //const [messageAPI, messageContext] = message.useMessage();
  const [requestType, setRequestType] = useState<string>();
  const [schemaData, setSchemaData] = useState<AsyncTask<Schema[], AppError>>({
    status: "pending",
  });
  const [userProfileData, setUserProfileData] = useState<AsyncTask<UserDetails[], AppError>>({
    status: "pending",
  });
  const env = useEnvContext();
  const { userDID } = useUserContext();
  // const userDID = localStorage.getItem("userId");
  const schemaList = isAsyncTaskDataAvailable(schemaData) ? schemaData.data : [];
  const profileList = isAsyncTaskDataAvailable(userProfileData) ? [userProfileData.data] : [];
  useEffect(() => {
    const getSchemas = async () => {
      const response = await getAllSchema({
        env,
      });
      if (response.success) {
        setSchemaData({
          data: response.data.successful,
          status: "successful",
        });
        notifyParseErrors(response.data.failed);
      } else {
        if (!isAbortedError(response.error)) {
          setSchemaData({ error: response.error, status: "failed" });
        }
      }

      // setschemaData(response.data.successful);
    };

    getSchemas().catch((e) => {
      console.error("An error occurred:", e);
    });
    const getUserDetails = async () => {
      try {
        const response = await getUser({
          env,
          userDID,
        });

        if (response.success) {
          setUserProfileData({
            data: response.data,
            status: "successful",
          });
          notifyParseErrors(response.data.failed);
        } else {
          if (!isAbortedError(response.error)) {
            setUserProfileData({ error: response.error, status: "failed" });
          }
        }
        // setUserProfileData(response.data);
      } catch (e) {
        console.error("An error occurred:", e);
      }
    };
    getUserDetails().catch((e) => {
      console.error("An error occurred:", e);
    });
  }, [env, userDID]);

  const handleFormSubmit = async (values: FormData) => {
    const schema = schemaList.find((item) => item.type === values.schemaID);
    console.log(values);

    const payload = {
      Age: values.age,
      ProofId: values.adhaarID,
      ProofType: "Adhar",
      RequestType: "GenerateNewVC",
      RoleType: "Individual",
      schemaID: schema?.id,
      Source: "Manual",
      userDID: userDID,
    };
    await requestVC({
      env,
      payload,
    }).then(void navigate(generatePath(ROUTES.request.path)));
  };

  return (
    <>
      {/* {messageContext} */}

      <SiderLayoutContent
        description="A request is issued with assigned attribute values and can be issued directly to identifier."
        showBackButton
        showDivider
        title={CREATE_REQUEST}
      >
        <Card className="issue-credential-card" title="Create Request">
          <Form
            layout="vertical"
            // eslint-disable-next-line
            onFinish={handleFormSubmit}
          >
            <Form.Item
              label="Select Crendential Type"
              name="schemaID"
              rules={[{ message: VALUE_REQUIRED, required: true }]}
            >
              <Select
                className="full-width"
                onChange={(id: string) => {
                  setRequestType(id);
                }}
                placeholder={SCHEMA_TYPE}
              >
                {schemaList.map((value: Schema, key: number) => (
                  <Select.Option key={key} value={value.type}>
                    {value.type}
                  </Select.Option>
                ))}
              </Select>
            </Form.Item>
            {(requestType === "KYCAgeCredentialAadhar" ||
              requestType === "ValidCredentialAadhar") && (
              <div>
                <Form.Item
                  label="Aadhaar Number"
                  name="adhaarID"
                  rules={[{ message: VALUE_REQUIRED, required: true }]}
                >
                  {/* eslint-disable */}
                  <Input defaultValue={profileList?.[0]?.adhar} placeholder="Adhaar Number" />
                  {/* eslint-enable */}
                </Form.Item>
              </div>
            )}
            {(requestType === "KYCAgeCredentialPAN" || requestType === "ValidCredentialPAN") && (
              <div>
                <Form.Item
                  label="PAN"
                  name="panID"
                  rules={[{ message: VALUE_REQUIRED, required: true }]}
                >
                  {/* eslint-disable */}
                  <Input defaultValue={profileList?.[0]?.PAN} placeholder="PAN" />
                  {/* eslint-enable */}
                </Form.Item>
              </div>
            )}
            {/* {requestType && (
              <div>
                <Form.Item
                  label="Request Type"
                  name="requestType"
                  rules={[{ message: VALUE_REQUIRED, required: true }]}
                >
                  <Input placeholder="Schema Type" />
                </Form.Item>
              </div>
            )} */}
            {(requestType === "KYCAgeCredentialAadhar" ||
              requestType === "KYCAgeCredentialPAN") && (
              <div>
                <Form.Item label="Age" name="age" rules={[{ message: VALUE_REQUIRED }]}>
                  <Input defaultValue="18" placeholder="Age" style={{ color: "#868686" }} />
                </Form.Item>
              </div>
            )}
            {requestType === "KYBGSTINCredentials" && (
              <div>
                <Form.Item
                  label="GSTIN"
                  name="gstin"
                  rules={[{ message: VALUE_REQUIRED, required: true }]}
                >
                  {/* eslint-disable */}
                  <Input defaultValue={profileList?.[0]?.gstin} placeholder="GSTIN" />
                  {/* eslint-enable */}
                </Form.Item>
              </div>
            )}
            <Button htmlType="submit" type="primary">
              {SUBMIT}
            </Button>
          </Form>
        </Card>
      </SiderLayoutContent>
    </>
  );
}
