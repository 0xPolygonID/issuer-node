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
import { AppError } from "src/domain";
import { FormValue, UserDetails } from "src/domain/user";
import { AsyncTask } from "src/utils/async";
import { isAbortedError } from "src/utils/browser";
import { PROFILE, PROFILE_DETAILS, VALUE_REQUIRED } from "src/utils/constants";
import { notifyParseErrors } from "src/utils/error";

export function Profile() {
  const { fullName, gmail, userDID, userType } = useUserContext();
  // const userDID = localStorage.getItem("userId");
  console.log(userDID);

  const [openModal, setOpenModal] = useState<boolean>(false);
  const [userProfileData, setUserProfileData] = useState<AsyncTask<UserDetails[], AppError>>({
    status: "pending",
  });
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
      .then(async (values: FormValue) => {
        const updatePayload = {
          Address: values.address,
          Adhar: values.Aadhar,
          DOB: values.dob,
          DocumentationSource: "manual",
          Gmail: "test@gmail.com",
          Gstin: values.gst,
          ID: userDID,
          Name: "test",
          Owner: values?.owner,
          PAN: values.PAN,
          PhoneNumber: values.mobile,
        };
        console.log("---", updatePayload);

        try {
          const userDetails = await updateUser({
            env,
            updatePayload,
          });

          if (userDetails.success) {
            localStorage.setItem("profile", "true");
            void messageAPI.success("Profile Updated");
            setOpenModal(false);
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
        const response = await getUser({
          env,
          userDID,
        });
        if (response.success) {
          setUserProfileData({
            data: response.data.successful,
            status: "successful",
          });
          notifyParseErrors(response.data.failed);
        } else {
          if (!isAbortedError(response.error)) {
            setUserProfileData({ error: response.error, status: "failed" });
          }
        }

        // setUserProfileData(response.data);
      };
      getUserDetails().catch((e) => {
        console.error("An error occurred:", e);
      });
    }
  }, [ProfileStatus, userDID, env]);
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
                  <Typography.Text>{userProfileData?.phoneNumber}</Typography.Text>
                </Row>
              </div>
            </Col>
            <Col span={6}>
              <Card style={{ height: 400, width: 600 }} title={PROFILE_DETAILS}>
                <Row>
                  <Typography.Text strong>UDID:</Typography.Text>
                  <Typography.Text>{userDID || userProfileData?.id}</Typography.Text>
                </Row>
                <Row>
                  <Typography.Text strong>Address</Typography.Text>
                  <Typography.Text>: {userProfileData?.address}</Typography.Text>
                </Row>
                <Row>
                  <Typography.Text strong>Adhaar Number</Typography.Text>
                  <Typography.Text style={{ marginRight: 10 }}>
                    : {userProfileData?.adhar}
                  </Typography.Text>
                  <UploadDoc />
                </Row>
                <Row>
                  <Typography.Text strong>PAN</Typography.Text>
                  <Typography.Text style={{ marginRight: 10 }}>
                    : {userProfileData?.PAN}
                  </Typography.Text>
                  <UploadDoc />
                </Row>
                <Row>
                  <Typography.Text strong>DOB</Typography.Text>
                  <Typography.Text>: {userProfileData?.dob}</Typography.Text>
                </Row>
              </Card>
            </Col>
          </Row>
        </Space>
      </SiderLayoutContent>
      <Modal onCancel={handleCancel} onOk={handleOk} open={openModal} title="Update Profile">
        <Form form={form} layout="vertical">
          <Form.Item
            label="DOB"
            name="dob"
            required
            rules={[{ message: VALUE_REQUIRED, required: true }]}
          >
            <Input placeholder="DOB" style={{ color: "#868686" }} />
          </Form.Item>
          <Form.Item
            label="Aadhar Number"
            name="Aadhar"
            required
            rules={[{ message: VALUE_REQUIRED, required: true }]}
          >
            <Input placeholder="Aadhar Number" style={{ color: "#868686" }} />
          </Form.Item>
          <Form.Item
            label="PAN"
            name="PAN"
            required
            rules={[{ message: VALUE_REQUIRED, required: true }]}
          >
            <Input placeholder="PAN" style={{ color: "#868686" }} />
          </Form.Item>
          <Form.Item
            label="User Type"
            name="request"
            required
            rules={[{ message: VALUE_REQUIRED, required: true }]}
          >
            <Input
              defaultValue={userType}
              placeholder="Request Type"
              style={{ color: "#868686" }}
            />
          </Form.Item>
          {userType !== "Individual" && (
            <Form.Item
              label="Owner"
              name="owner"
              required
              rules={[{ message: VALUE_REQUIRED, required: true }]}
            >
              <Input placeholder="Owner" style={{ color: "#868686" }} />
            </Form.Item>
          )}
          {userType !== "Individual" && (
            <Form.Item
              label="GSTIN"
              name="gst"
              required
              rules={[{ message: VALUE_REQUIRED, required: true }]}
            >
              <Input placeholder="GSTIN" style={{ color: "#868686" }} />
            </Form.Item>
          )}
          <Form.Item
            label="Address"
            name="address"
            required
            rules={[{ message: VALUE_REQUIRED, required: true }]}
          >
            <Input placeholder="Address" required style={{ color: "#868686" }} />
          </Form.Item>
          <Form.Item
            label="Mobile Number"
            name="mobile"
            required
            rules={[{ message: VALUE_REQUIRED, required: true }]}
          >
            <Input placeholder="Mobile Number" style={{ color: "#868686" }} />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
