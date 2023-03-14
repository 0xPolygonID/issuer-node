import {
  Button,
  Card,
  Col,
  DatePicker,
  Form,
  InputNumber,
  Radio,
  Row,
  Space,
  Tag,
  TimePicker,
  Typography,
} from "antd";
import dayjs from "dayjs";
import { useState } from "react";

import { parseClaimLinkExpirationDate } from "src/adapters/parsers/forms";
import { ReactComponent as IconBack } from "src/assets/icons/arrow-narrow-left.svg";
import { ReactComponent as IconRight } from "src/assets/icons/arrow-narrow-right.svg";
import { FORM_LABEL } from "src/utils/constants";

export type IssuanceMethod =
  | {
      type: "directIssue";
    }
  | {
      claimLinkExpirationDate?: dayjs.Dayjs;
      claimLinkExpirationTime?: dayjs.Dayjs;
      limitedClaims?: string;
      type: "claimLink";
    };

export function SetIssuanceMethod({
  claimExpirationDate,
  initialValues,
  isClaimLoading,
  onStepBack,
  onSubmit,
}: {
  claimExpirationDate?: dayjs.Dayjs;
  initialValues: IssuanceMethod;
  isClaimLoading: boolean;
  onStepBack: (values: IssuanceMethod) => void;
  onSubmit: (values: IssuanceMethod) => void;
}) {
  const [issuanceMethod, setIssuanceMethod] = useState<IssuanceMethod>(initialValues);
  const isDirectIssue = issuanceMethod.type === "directIssue";

  return (
    <Card className="claiming-card" title="Select a way to issue claims">
      <Form
        initialValues={initialValues}
        layout="vertical"
        name="issueClaimMethod"
        onFinish={onSubmit}
        onValuesChange={(changedValues, allValues) => {
          const parsedClaimLinkExpirationDate =
            parseClaimLinkExpirationDate.safeParse(changedValues);

          if (
            allValues.type === "claimLink" &&
            parsedClaimLinkExpirationDate.success &&
            (parsedClaimLinkExpirationDate.data.claimLinkExpirationDate === null ||
              (dayjs().isSame(parsedClaimLinkExpirationDate.data.claimLinkExpirationDate, "day") &&
                dayjs().isAfter(allValues.claimLinkExpirationTime)))
          ) {
            setIssuanceMethod({ ...allValues, claimLinkExpirationTime: undefined });
          } else {
            setIssuanceMethod(allValues);
          }
        }}
        requiredMark={false}
        validateTrigger="onBlur"
      >
        <Form.Item name="type" rules={[{ message: "Value required", required: true }]}>
          <Radio.Group name="type" style={{ width: "100%" }}>
            <Space direction="vertical">
              <Card className={`${isDirectIssue ? "selected" : ""} disabled`}>
                <Radio disabled value="directIssue">
                  <Space direction="vertical">
                    <Typography.Text>
                      Direct issue <Tag>Coming soon</Tag>
                    </Typography.Text>

                    <Typography.Text type="secondary">
                      Issue claims directly using a known identifier - connections with your
                      organization or establish connection with new identifiers.
                    </Typography.Text>
                  </Space>
                </Radio>
              </Card>

              <Card className={issuanceMethod.type === "claimLink" ? "selected" : ""}>
                <Space direction="vertical" size="large">
                  <Radio value="claimLink">
                    <Space direction="vertical">
                      <Typography.Text>{FORM_LABEL.CLAIM_LINK}</Typography.Text>

                      <Typography.Text type="secondary">
                        Anyone can access the claim with this link. You can deactivate it at any
                        time.
                      </Typography.Text>
                    </Space>
                  </Radio>

                  <Space direction="horizontal" size="large" style={{ paddingLeft: 28 }}>
                    <Space align="end" direction="horizontal">
                      <Form.Item
                        help="Optional"
                        label={FORM_LABEL.LINK_VALIDITY}
                        name="claimLinkExpirationDate"
                      >
                        <DatePicker
                          disabled={isDirectIssue}
                          disabledDate={(current) =>
                            current < dayjs().startOf("day") ||
                            (claimExpirationDate !== undefined &&
                              current.isAfter(claimExpirationDate))
                          }
                        />
                      </Form.Item>

                      <Form.Item
                        getValueProps={() => {
                          return {
                            claimLinkExpirationTime:
                              issuanceMethod.type === "claimLink" &&
                              issuanceMethod.claimLinkExpirationTime,
                          };
                        }}
                        name="claimLinkExpirationTime"
                      >
                        <TimePicker
                          disabled={isDirectIssue}
                          disabledTime={() => {
                            const now = dayjs();

                            if (
                              issuanceMethod.type === "claimLink" &&
                              now.isSame(issuanceMethod.claimLinkExpirationDate, "day")
                            ) {
                              return {
                                disabledHours: () => [...Array(now.hour()).keys()],
                                disabledMinutes: (hour) => {
                                  return now.hour() === hour
                                    ? [...Array(now.minute() + 1).keys()]
                                    : hour < 0
                                    ? [...Array(60).keys()]
                                    : [];
                                },
                              };
                            } else {
                              return {};
                            }
                          }}
                          format="HH:mm"
                          hideDisabledOptions
                          minuteStep={5}
                          showNow={false}
                          value={
                            issuanceMethod.type === "claimLink"
                              ? issuanceMethod.claimLinkExpirationTime
                              : undefined
                          }
                        />
                      </Form.Item>
                    </Space>

                    <Form.Item help="Optional" label="Set maximum issuance" name="limitedClaims">
                      <InputNumber
                        className="full-width"
                        disabled={isDirectIssue}
                        min={1}
                        placeholder="e.g 1000"
                        size="large"
                        type="number"
                      />
                    </Form.Item>
                  </Space>
                </Space>
              </Card>
            </Space>
          </Radio.Group>
        </Form.Item>

        <Row gutter={8} justify="end">
          <Col>
            <Button
              icon={<IconBack />}
              loading={isClaimLoading}
              onClick={() => {
                onStepBack(issuanceMethod);
              }}
            >
              Previous step
            </Button>
          </Col>

          <Col>
            <Button htmlType="submit" loading={isClaimLoading} type="primary">
              Create claim link
              <IconRight />
            </Button>
          </Col>
        </Row>
      </Form>
    </Card>
  );
}
