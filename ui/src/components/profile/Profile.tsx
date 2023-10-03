import { Card, Col, Divider, Image, Row, Space, Typography } from "antd";

import { UploadDoc } from "../shared/Upload";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";

import { PROFILE, PROFILE_DETAILS } from "src/utils/constants";

export function Profile() {
  const src = "https://zos.alipayobjects.com/rmsportal/jkjgkEfvpUPVyRjUImniVslZfWPnJuuZ.png";
  return (
    <SiderLayoutContent title={PROFILE}>
      <Divider />
      <Space className="d-flex" direction="vertical">
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
                <Typography.Text>Roshni</Typography.Text>
              </Row>
              <Row>
                <Typography.Text>roshni@chaincodeconsulting.com</Typography.Text>
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
                <Typography.Text>: XYZ</Typography.Text>
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
  );
}
