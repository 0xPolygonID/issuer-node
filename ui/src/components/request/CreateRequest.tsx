import { Card, Form, Input, Select } from "antd";
import { useState } from "react";

import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";

import { CREATE_REQUEST, SCHEMA_TYPE, VALUE_REQUIRED } from "src/utils/constants";

const dropdownList = [
  "KYCAgeCredentialAadhar",
  "ValidCredentialAadhar",
  "KYCAgeCredentialPAN",
  "ValidCredentialPAN",
  "KYBGSTINCredentials",
];

export function CreateRequest() {
  //const [messageAPI, messageContext] = message.useMessage();
  const [requestType, setRequestType] = useState<string>();

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
          <Form layout="vertical">
            <Form.Item
              label="Select schema type"
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
                {dropdownList.map((value, key) => (
                  <Select.Option key={key} value={value}>
                    {value}
                  </Select.Option>
                ))}
              </Select>
            </Form.Item>
            {(requestType === "KYCAgeCredentialAadhar" ||
              requestType === "ValidCredentialAadhar") && (
              <div>
                <Form.Item
                  label="Adhaar"
                  name="adhaarID"
                  rules={[{ message: VALUE_REQUIRED, required: true }]}
                >
                  <Input placeholder="Adhaar Number" />
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
                  <Input placeholder="Adhaar Number" />
                </Form.Item>
              </div>
            )}
            {requestType && (
              <div>
                <Form.Item
                  label="Request Type"
                  name="requestType"
                  rules={[{ message: VALUE_REQUIRED, required: true }]}
                >
                  <Input placeholder="Schema Type" />
                </Form.Item>
              </div>
            )}
            {(requestType === "KYCAgeCredentialAadhar" ||
              requestType === "KYCAgeCredentialPAN") && (
              <div>
                <Form.Item
                  label="Age"
                  name="age"
                  rules={[{ message: VALUE_REQUIRED, required: true }]}
                >
                  <Input defaultValue={18} placeholder="Age" />
                </Form.Item>
              </div>
            )}
          </Form>
        </Card>
      </SiderLayoutContent>
    </>
  );
}
