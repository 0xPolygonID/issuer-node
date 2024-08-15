import { Button, Col, Divider, Form, Input, Row, Select, Typography } from "antd";
import { useState } from "react";
import { IssuerFormData, issuerFormDataParser } from "src/adapters/parsers/view";
import {
  AuthBJJCredentialStatus,
  Blockchain,
  IssuerType,
  Method,
  PolygonNetwork,
  PrivadoNetwork,
} from "src/domain";
import { VALUE_REQUIRED } from "src/utils/constants";

export function IssuerForm({
  initialValues,
  onSubmit,
  submitBtnText,
}: {
  initialValues: IssuerFormData;
  onSubmit: (formValues: IssuerFormData) => void;
  submitBtnText: string;
}) {
  const [form] = Form.useForm<IssuerFormData>();
  const [formData, setFormData] = useState<IssuerFormData>(initialValues);

  const showCredentialStatusField: boolean = formData.type === IssuerType.BJJ;

  return (
    <Form
      form={form}
      initialValues={formData}
      layout="vertical"
      onFinish={onSubmit}
      onValuesChange={(changedValue: Partial<IssuerFormData>, allValues) => {
        const updatedFormData = { ...allValues };

        if (updatedFormData.type === IssuerType.BJJ && !updatedFormData.authBJJCredentialStatus) {
          updatedFormData.authBJJCredentialStatus =
            AuthBJJCredentialStatus.Iden3OnchainSparseMerkleTreeProof2023;
        }

        if (changedValue.blockchain) {
          updatedFormData.network =
            changedValue.blockchain === Blockchain.polygon
              ? PolygonNetwork.mainnet
              : PrivadoNetwork.main;
        }

        const parsedIssuerFormData = issuerFormDataParser.safeParse(updatedFormData);

        if (parsedIssuerFormData.success) {
          setFormData(parsedIssuerFormData.data);
          form.setFieldsValue(parsedIssuerFormData.data);
        }
      }}
    >
      <Form.Item>
        <Form.Item
          label="Issuer name"
          name="displayName"
          rules={[{ message: VALUE_REQUIRED, required: true }]}
        >
          <Input placeholder="Enter name" />
        </Form.Item>
        <Typography.Text type="secondary">
          Give your issuer a descriptive name, e.g. “Age credential testing”. This name is only seen
          locally.
        </Typography.Text>
      </Form.Item>

      <Form.Item>
        <Form.Item
          label="Method"
          name="method"
          rules={[{ message: VALUE_REQUIRED, required: true }]}
        >
          <Select
            className="full-width"
            disabled={Object.values(Method).length < 2}
            placeholder="Method"
          >
            {Object.values(Method).map((method) => (
              <Select.Option key={method} value={method}>
                {method}
              </Select.Option>
            ))}
          </Select>
        </Form.Item>
        <Typography.Text type="secondary">
          The protocol or system used to create, resolve, and manage the DID.
        </Typography.Text>
      </Form.Item>

      <Form.Item
        label="Blockchain"
        name="blockchain"
        rules={[{ message: VALUE_REQUIRED, required: true }]}
      >
        <Select className="full-width" placeholder="Type">
          {Object.values(Blockchain).map((blockchain) => (
            <Select.Option key={blockchain} value={blockchain}>
              {blockchain}
            </Select.Option>
          ))}
        </Select>
      </Form.Item>

      <Form.Item
        label="Network"
        name="network"
        rules={[{ message: VALUE_REQUIRED, required: true }]}
      >
        <Select className="full-width" placeholder="Network">
          {formData.blockchain === Blockchain.polygon &&
            Object.values(PolygonNetwork).map((network) => (
              <Select.Option key={network} value={network}>
                {network}
              </Select.Option>
            ))}
          {formData.blockchain === Blockchain.privado &&
            Object.values(PrivadoNetwork).map((network) => (
              <Select.Option key={network} value={network}>
                {network}
              </Select.Option>
            ))}
          {}
        </Select>
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
        <Form.Item>
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
          <Typography.Text type="secondary">
            Credential status * Iden3OnchainSparskeMerkleTreeProof2023 @olivia Identity signing
            key&apos;s credential status is checked by clients to generate zero-knowledge proofs
            using signed credentials.
          </Typography.Text>
        </Form.Item>
      )}
      <>
        <Divider />
        <Row gutter={[8, 8]} justify="end">
          <Col>
            <Button htmlType="submit" type="primary">
              {submitBtnText}
            </Button>
          </Col>
        </Row>
      </>
    </Form>
  );
}
