import {
  AutoComplete,
  Button,
  Card,
  DatePicker,
  Form,
  Input,
  InputNumber,
  Radio,
  Row,
  Space,
  TimePicker,
  Typography,
} from "antd";
import dayjs from "dayjs";
import { useCallback, useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";

import { APIError } from "src/adapters/api";
import { getConnections } from "src/adapters/api/connections";
import { IssuanceMethodFormData, linkExpirationDateParser } from "src/adapters/parsers/forms";
import { ReactComponent as IconRight } from "src/assets/icons/arrow-narrow-right.svg";
import { useEnvContext } from "src/contexts/Env";
import { Connection } from "src/domain";
import { AsyncTask, isAsyncTaskDataAvailable } from "src/utils/async";
import { makeRequestAbortable } from "src/utils/browser";
import { ACCESSIBLE_UNTIL, CREDENTIAL_LINK, DID_SEARCH_PARAM } from "src/utils/constants";

export function IssuanceMethodForm({
  initialValues,
  onSubmit,
}: {
  initialValues: IssuanceMethodFormData;
  onSubmit: (values: IssuanceMethodFormData) => void;
}) {
  const env = useEnvContext();
  const [searchParams] = useSearchParams();

  const [issuanceMethod, setIssuanceMethod] = useState<IssuanceMethodFormData>(
    initialValues.type === "directIssue"
      ? {
          ...initialValues,
          did: searchParams.get(DID_SEARCH_PARAM) || "",
        }
      : initialValues
  );
  const [connections, setConnections] = useState<AsyncTask<Connection[], APIError>>({
    status: "pending",
  });

  const isLinkIssue = issuanceMethod.type === "credentialLink";
  const isDirectIssue = issuanceMethod.type === "directIssue";

  const isNextButtonDisabled = isDirectIssue && issuanceMethod.did === "";

  const isConnectedSuffixVisible =
    isDirectIssue &&
    isAsyncTaskDataAvailable(connections) &&
    connections.data.find((connection) => connection.userID === issuanceMethod.did) !== undefined;

  const fetchConnections = useCallback(
    async (signal: AbortSignal) => {
      const response = await getConnections({ credentials: false, env, signal });

      if (response.isSuccessful) {
        setConnections({ data: response.data, status: "successful" });
      } else {
        setConnections({ error: response.error, status: "failed" });
      }
    },
    [env]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchConnections);

    return () => aborter();
  }, [fetchConnections]);

  return (
    <Card className="issue-credential-card" title="Choose how to issue credential">
      <Form
        initialValues={issuanceMethod}
        layout="vertical"
        name="issueCredentialMethod"
        onFinish={onSubmit}
        onValuesChange={(changedValues, allValues) => {
          const parsedLinkExpirationDate = linkExpirationDateParser.safeParse(changedValues);
          if (
            allValues.type === "credentialLink" &&
            parsedLinkExpirationDate.success &&
            (parsedLinkExpirationDate.data.linkExpirationDate === null ||
              (dayjs().isSame(parsedLinkExpirationDate.data.linkExpirationDate, "day") &&
                dayjs().isAfter(allValues.linkExpirationTime)))
          ) {
            setIssuanceMethod({ ...allValues, linkExpirationTime: undefined });
          } else {
            setIssuanceMethod(allValues);
          }
        }}
        requiredMark={false}
        validateTrigger="onBlur"
      >
        <Form.Item name="type" rules={[{ message: "Value required", required: true }]}>
          <Radio.Group className="full-width" name="type">
            <Space direction="vertical">
              <Card className={`${isDirectIssue ? "selected" : ""}`}>
                <Radio value="directIssue">
                  <Space direction="vertical">
                    <Typography.Text>Direct issue</Typography.Text>

                    <Typography.Text type="secondary">
                      Issue credentials directly using a known identifier - connections with your
                      organization or establish connection with new identifiers.
                    </Typography.Text>
                  </Space>
                </Radio>

                <Form.Item
                  label="Select connection/Paste identifier"
                  name="did"
                  required
                  style={{ paddingLeft: 28, paddingTop: 16 }}
                >
                  <AutoComplete
                    disabled={isLinkIssue}
                    filterOption={(inputValue, option) =>
                      option !== undefined
                        ? option.value.toUpperCase().indexOf(inputValue.toUpperCase()) !== -1
                        : false
                    }
                    options={
                      isAsyncTaskDataAvailable(connections)
                        ? connections.data.map(({ userID }) => {
                            const network = userID.split(":").splice(0, 4).join(":");
                            const did = userID.split(":").pop();

                            if (did) {
                              return {
                                label: `${network}:${did.slice(0, 6)}...${did.slice(-6)}`,
                                value: userID,
                              };
                            } else {
                              return { label: userID, value: userID };
                            }
                          })
                        : undefined
                    }
                  >
                    <Input
                      className={isConnectedSuffixVisible ? undefined : "hidden-suffix"}
                      placeholder="Select or paste"
                      suffix={<Typography.Text type="secondary">Connected</Typography.Text>}
                    />
                  </AutoComplete>
                </Form.Item>
              </Card>

              <Card className={issuanceMethod.type === "credentialLink" ? "selected" : ""}>
                <Space direction="vertical" size="large">
                  <Radio value="credentialLink">
                    <Space direction="vertical">
                      <Typography.Text>{CREDENTIAL_LINK}</Typography.Text>

                      <Typography.Text type="secondary">
                        Anyone can access the credential with this link. You can deactivate it at
                        any time.
                      </Typography.Text>
                    </Space>
                  </Radio>

                  <Space direction="horizontal" size="large" style={{ paddingLeft: 28 }}>
                    <Space align="end" direction="horizontal">
                      <Form.Item help="Optional" label={ACCESSIBLE_UNTIL} name="linkExpirationDate">
                        <DatePicker
                          disabled={isDirectIssue}
                          disabledDate={(current) => current < dayjs().startOf("day")}
                        />
                      </Form.Item>

                      <Form.Item
                        getValueProps={() => {
                          return {
                            linkExpirationTime:
                              issuanceMethod.type === "credentialLink" &&
                              issuanceMethod.linkExpirationTime,
                          };
                        }}
                        name="linkExpirationTime"
                      >
                        <TimePicker
                          disabled={isDirectIssue}
                          disabledTime={() => {
                            const now = dayjs();

                            if (
                              issuanceMethod.type === "credentialLink" &&
                              now.isSame(issuanceMethod.linkExpirationDate, "day")
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
                            issuanceMethod.type === "credentialLink"
                              ? issuanceMethod.linkExpirationTime
                              : undefined
                          }
                        />
                      </Form.Item>
                    </Space>

                    <Form.Item
                      help="Optional"
                      label="Set maximum issuance"
                      name="linkMaximumIssuance"
                    >
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
          <Button disabled={isNextButtonDisabled} htmlType="submit" type="primary">
            Next step <IconRight />
          </Button>
        </Row>
      </Form>
    </Card>
  );
}
