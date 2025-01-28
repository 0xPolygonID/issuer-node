import { Card, Flex, Space, Typography } from "antd";
import { useCallback, useEffect, useState } from "react";
import { generatePath, useParams } from "react-router-dom";

import { useIdentityContext } from "../../contexts/Identity";
import { getPaymentOptions, getPaymentRequest } from "src/adapters/api/payments";
import { notifyErrors } from "src/adapters/parsers";
import { JSONHighlighter } from "src/components/schemas/JSONHighlighter";
import { Detail } from "src/components/shared/Detail";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { AppError, PaymentOption, PaymentRequest as PaymentRequestType } from "src/domain";
import { ROUTES } from "src/routes";
import {
  AsyncTask,
  hasAsyncTaskFailed,
  isAsyncTaskDataAvailable,
  isAsyncTaskStarting,
} from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { PAYMENT_REQUESTS_DETAILS } from "src/utils/constants";
import { formatDate } from "src/utils/forms";

export function PaymentRequest() {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();

  const [paymentRequest, setPaymentRequest] = useState<AsyncTask<PaymentRequestType, AppError>>({
    status: "pending",
  });

  const [paymentOptions, setPaymentOptions] = useState<AsyncTask<PaymentOption[], AppError>>({
    status: "pending",
  });

  const { paymentRequestID } = useParams();

  const fetchPaymentRequest = useCallback(
    async (signal?: AbortSignal) => {
      if (paymentRequestID) {
        setPaymentRequest({ status: "loading" });

        const response = await getPaymentRequest({
          env,
          identifier,
          paymentRequestID,
          signal,
        });

        if (response.success) {
          setPaymentRequest({ data: response.data, status: "successful" });
        } else {
          if (!isAbortedError(response.error)) {
            setPaymentRequest({ error: response.error, status: "failed" });
          }
        }
      }
    },
    [env, paymentRequestID, identifier]
  );

  const fetchPaymentOptions = useCallback(
    async (signal?: AbortSignal) => {
      setPaymentOptions((previousPaymentOptions) =>
        isAsyncTaskDataAvailable(previousPaymentOptions)
          ? { data: previousPaymentOptions.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await getPaymentOptions({
        env,
        identifier,
        params: {},
        signal,
      });
      if (response.success) {
        setPaymentOptions({
          data: response.data.items.successful,
          status: "successful",
        });

        void notifyErrors(response.data.items.failed);
      } else {
        if (!isAbortedError(response.error)) {
          setPaymentOptions({ error: response.error, status: "failed" });
        }
      }
    },
    [env, identifier]
  );
  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchPaymentRequest);

    return aborter;
  }, [fetchPaymentRequest]);

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchPaymentOptions);

    return aborter;
  }, [fetchPaymentOptions]);

  if (!paymentRequestID) {
    return <ErrorResult error="No payment request provided." />;
  }

  return (
    <SiderLayoutContent
      description="View payment request details"
      showBackButton
      showDivider
      title={PAYMENT_REQUESTS_DETAILS}
    >
      {(() => {
        if (hasAsyncTaskFailed(paymentRequest)) {
          return (
            <Card className="centered">
              <ErrorResult
                error={[
                  "An error occurred while downloading a payment request from the API:",
                  paymentRequest.error.message,
                ].join("\n")}
              />
            </Card>
          );
        } else if (hasAsyncTaskFailed(paymentOptions)) {
          return (
            <Card className="centered">
              <ErrorResult
                error={[
                  "An error occurred while downloading a payment options from the API:",
                  paymentOptions.error.message,
                ].join("\n")}
              />
            </Card>
          );
        } else if (isAsyncTaskStarting(paymentRequest) || isAsyncTaskStarting(paymentOptions)) {
          return (
            <Card className="centered">
              <LoadingResult />
            </Card>
          );
        } else {
          const paymentOptionName = paymentOptions.data.find(
            ({ id }) => id === paymentRequest.data.paymentOptionID
          )?.name;

          return (
            <>
              <Card
                className="centered"
                title={
                  <Flex align="center" gap={8} justify="space-between">
                    <Typography.Text style={{ fontWeight: 600 }}>
                      {paymentRequest.data.id}
                    </Typography.Text>
                  </Flex>
                }
              >
                <Flex gap={24} vertical>
                  <Card className="background-grey">
                    <Space direction="vertical">
                      <Detail
                        copyable
                        ellipsisPosition={5}
                        label="User DID"
                        text={paymentRequest.data.userDID}
                      />
                      <Detail
                        label="Created date"
                        text={formatDate(paymentRequest.data.createdAt)}
                      />
                      <Detail
                        label="Modified date"
                        text={formatDate(paymentRequest.data.modifiedAt)}
                      />
                      <Detail label="Status" text={paymentRequest.data.status} />
                      <Detail
                        href={generatePath(ROUTES.paymentOptionDetails.path, {
                          paymentOptionID: paymentRequest.data.paymentOptionID,
                        })}
                        label="Payment option"
                        text={paymentOptionName || paymentRequest.data.paymentOptionID}
                      />
                    </Space>
                  </Card>

                  <Flex gap={8} vertical>
                    <Typography.Text>Payments:</Typography.Text>
                    <JSONHighlighter json={paymentRequest.data.payments} />
                  </Flex>
                </Flex>
              </Card>
            </>
          );
        }
      })()}
    </SiderLayoutContent>
  );
}
