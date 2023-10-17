import { Button, Card, Form, Input, Select } from "antd";
import { useState } from "react";

import { SCHEMA_TYPE, SUBMIT, VALUE_REQUIRED } from "src/utils/constants";

const dropdownList = [
  "KYCAgeCredentialAadhar",
  "ValidCredentialAadhar",
  "KYCAgeCredentialPAN",
  "ValidCredentialPAN",
  "KYBGSTINCredentials",
];

export function ProfileUpdateModal() {
  const [requestType, setRequestType] = useState<string>();

  return (
    <>
      <Card className="issue-credential-card" title="Create Request">
        <Form layout="vertical">
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
                label="Aadhaar Number"
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
          {(requestType === "KYCAgeCredentialAadhar" || requestType === "KYCAgeCredentialPAN") && (
            <div>
              <Form.Item label="Age" name="age" rules={[{ message: VALUE_REQUIRED }]}>
                <Input defaultValue={18} placeholder="Age" readOnly style={{ color: "#868686" }} />
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
                <Input placeholder="GSTIN" readOnly />
              </Form.Item>
            </div>
          )}
          <Button key="submit" type="primary">
            {SUBMIT}
          </Button>
        </Form>
      </Card>
    </>
  );
}
