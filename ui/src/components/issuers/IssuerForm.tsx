import { Button, Col, Divider, Form, Input, Row, Select } from "antd";
import { useState } from "react";
import { IssuerFormData } from "src/adapters/parsers/view";
import IconBack from "src/assets/icons/arrow-narrow-left.svg?react";
import { AuthBJJCredentialStatus, IssuerType } from "src/domain";
import { VALUE_REQUIRED } from "src/utils/constants";

export function IssuerForm({
  onBack,
  onSubmit,
}: {
  onBack: () => void;
  onSubmit: (formValues: IssuerFormData) => void;
}) {
  const [form] = Form.useForm<IssuerFormData>();
  const [showCredentialStatusField, setShowCredentialStatusField] = useState(false);

  const handleFormChange = (changedValues: Partial<IssuerFormData>) => {
    if (changedValues.type) {
      if (changedValues.type === IssuerType.BJJ) {
        setShowCredentialStatusField(true);
      } else {
        setShowCredentialStatusField(false);
      }
    }
  };

  return (
    <Form form={form} layout="vertical" onFinish={onSubmit} onValuesChange={handleFormChange}>
      <Form.Item label="Method" name="method" rules={[{ message: VALUE_REQUIRED, required: true }]}>
        <Input placeholder="Method" />
      </Form.Item>
      <Form.Item
        label="Blockchain"
        name="blockchain"
        rules={[{ message: VALUE_REQUIRED, required: true }]}
      >
        <Input placeholder="Blockchain" />
      </Form.Item>
      <Form.Item
        label="Network"
        name="network"
        rules={[{ message: VALUE_REQUIRED, required: true }]}
      >
        <Input placeholder="Network" />
      </Form.Item>
      <Form.Item label="Type" name="type" rules={[{ message: VALUE_REQUIRED, required: true }]}>
        <Select className="full-width" placeholder="Type">
          {Object.values(IssuerType).map((type) => (
            <Select.Option key={type} value={type}>
              {type}
            </Select.Option>
          ))}
        </Select>
      </Form.Item>

      {showCredentialStatusField && (
        <Form.Item
          label="Credential Status"
          name="authBJJCredentialStatus"
          rules={[{ message: VALUE_REQUIRED, required: true }]}
        >
          <Select className="full-width" placeholder="Credential Status">
            {Object.values(AuthBJJCredentialStatus).map((credentialStatus) => (
              <Select.Option key={credentialStatus} value={credentialStatus}>
                {credentialStatus}
              </Select.Option>
            ))}
          </Select>
        </Form.Item>
      )}
      <>
        <Divider />
        <Row gutter={[8, 8]} justify="end">
          <Col>
            <Button icon={<IconBack />} onClick={onBack} type="default">
              Back to List
            </Button>
          </Col>

          <Col>
            <Button htmlType="submit" type="primary">
              Add issuer
            </Button>
          </Col>
        </Row>
      </>
    </Form>
  );
}
