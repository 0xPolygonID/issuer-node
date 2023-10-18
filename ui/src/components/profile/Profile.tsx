import {
  Button,
  Card,
  Col,
  Divider,
  Form,
  Image,
  Input,
  Modal,
  Row,
  Space,
  Typography,
  message,
} from "antd";

import { useEffect, useState } from "react";
import { UploadDoc } from "../shared/Upload";
import { getUser, updateUser } from "src/adapters/api/user";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";

import { useEnvContext } from "src/contexts/Env";
import { useUserContext } from "src/contexts/UserDetails";
import { PROFILE, PROFILE_DETAILS, VALUE_REQUIRED } from "src/utils/constants";

export function Profile() {
  const { fullName, gmail, UserDID, userType } = useUserContext();

  const [openModal, setOpenModal] = useState<boolean>(false);
  const env = useEnvContext();
  const src = "https://zos.alipayobjects.com/rmsportal/jkjgkEfvpUPVyRjUImniVslZfWPnJuuZ.png";
  const [messageAPI, messageContext] = message.useMessage();
  const [form] = Form.useForm();
  const ProfileStatus = localStorage.getItem("profile");

  const handleCancel = () => {
    setOpenModal(false);
  };
  const handleOk = () => {
    form
      .validateFields()
      .then(async (values) => {
        console.log("Submitted", values);
        // const updatePayload = {
        //   ID: UserDID,
        //   Name: "test",
        //   Owner: values.Owner,
        //   Gmail: "test@gmail.com",
        //   Gstin: values.gst,
        //   Address: values.address,
        //   Adhar: values.Aadhar,
        //   PAN: values.PAN,
        //   DocumentationSource: "manual",
        // };
        try {
          const userDetails = await updateUser({
            env,
            UserDID,
          });

          if (userDetails.success) {
            localStorage.setItem("profile", "true");
            void messageAPI.success("Profile Updated");
          } else {
            void messageAPI.error("Wrong Credentials");
          }
        } catch (error) {
          // Handle the error, e.g., show an error message
          console.error("An error occurred:", error);
        }
      })
      .catch((e) => {
        console.error("An error occurred:", e);
      });
  };

  useEffect(() => {
    if (ProfileStatus === "true") {
      const getUserDetails = async () => {
        await getUser({
          env,
          UserDID,
        });
      };
      getUserDetails().catch((e) => {
        console.error("An error occurred:", e);
      });
    }
  }, [ProfileStatus, UserDID, env]);
  return (
    <>
      {messageContext}
      <SiderLayoutContent title={PROFILE}>
        <Divider />
        <Space className="d-flex" direction="vertical">
          <Button onClick={() => setOpenModal(true)} type="primary">
            Update
          </Button>
          <Row gutter={50}>
            <Col span={12}>
              <div
                style={{
                  alignItems: "center",
                  backgroundColor: "white",
                  border: "1px solid #f0f0f0",
                  borderRadius: "10px",
                  display: "flex",
                  flexDirection: "column",
                  height: 400,
                  justifyContent: "center",
                  textAlign: "center",
                  width: 600,
                }}
              >
                <Image src={src} style={{ borderRadius: 100, marginBottom: 10 }} width={200} />
                <Row>
                  <Typography.Text>{fullName}</Typography.Text>
                </Row>
                <Row>
                  <Typography.Text>{gmail}</Typography.Text>
                </Row>
                <Row>
                  <Typography.Text>7008714710</Typography.Text>
                </Row>
              </div>
            </Col>
            <Col span={6}>
              <Card style={{ height: 400, width: 600 }} title={PROFILE_DETAILS}>
                <Row>
                  <Typography.Text strong>UDID</Typography.Text>
                  <Typography.Text>: {UserDID}</Typography.Text>
                </Row>
                <Row>
                  <Typography.Text strong>Address</Typography.Text>
                  <Typography.Text>: College Road</Typography.Text>
                </Row>
                <Row>
                  <Typography.Text strong>Adhaar Number</Typography.Text>
                  <Typography.Text style={{ marginRight: 10 }}>: 7008714710</Typography.Text>
                  <UploadDoc />
                </Row>
                <Row>
                  <Typography.Text strong>PAN</Typography.Text>
                  <Typography.Text style={{ marginRight: 10 }}>: 7008714710</Typography.Text>
                  <UploadDoc />
                </Row>
                <Row>
                  <Typography.Text strong>Age</Typography.Text>
                  <Typography.Text>: 25</Typography.Text>
                </Row>
              </Card>
            </Col>
          </Row>
        </Space>
      </SiderLayoutContent>
      <Modal onCancel={handleCancel} onOk={handleOk} open={openModal} title="Update Profile">
        <Form form={form} layout="vertical">
          <Form.Item
            label="Age"
            name="age"
            required
            rules={[{ message: VALUE_REQUIRED, required: true }]}
          >
            <Input placeholder="Age" style={{ color: "#868686" }} />
          </Form.Item>
          <Form.Item
            label="Aadhar Number"
            name="Aadhar"
            required
            rules={[{ message: VALUE_REQUIRED, required: true }]}
          >
            <Input placeholder="Aadhar Number" readOnly style={{ color: "#868686" }} />
          </Form.Item>
          <Form.Item
            label="PAN"
            name="PAN"
            required
            rules={[{ message: VALUE_REQUIRED, required: true }]}
          >
            <Input placeholder="PAN" readOnly style={{ color: "#868686" }} />
          </Form.Item>
          <Form.Item
            label="Request- Type"
            name="request"
            required
            rules={[{ message: VALUE_REQUIRED, required: true }]}
          >
            <Input placeholder="Request Type" readOnly style={{ color: "#868686" }} />
          </Form.Item>
          {userType !== "Individual" && (
            <Form.Item
              label="Owner"
              name="owner"
              required
              rules={[{ message: VALUE_REQUIRED, required: true }]}
            >
              <Input placeholder="Owner" readOnly style={{ color: "#868686" }} />
            </Form.Item>
          )}
          {userType !== "Individual" && (
            <Form.Item
              label="GSTIN"
              name="gst"
              required
              rules={[{ message: VALUE_REQUIRED, required: true }]}
            >
              <Input placeholder="GSTIN" readOnly style={{ color: "#868686" }} />
            </Form.Item>
          )}
          <Form.Item
            label="Address"
            name="address"
            required
            rules={[{ message: VALUE_REQUIRED, required: true }]}
          >
            <Input placeholder="Address" readOnly required style={{ color: "#868686" }} />
          </Form.Item>
          <Form.Item
            label="Mobile Number"
            name="mobile"
            required
            rules={[{ message: VALUE_REQUIRED, required: true }]}
          >
            <Input placeholder="Mobile Number" readOnly style={{ color: "#868686" }} />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
