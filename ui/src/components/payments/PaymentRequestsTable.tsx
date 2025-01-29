import {
  Avatar,
  Card,
  Dropdown,
  Row,
  Space,
  Table,
  TableColumnsType,
  Tag,
  Tooltip,
  Typography,
} from "antd";
import { useCallback, useEffect, useState } from "react";
import { generatePath, useNavigate } from "react-router-dom";

import { PaymentRequestStatus } from "../../domain/payment";
import { getPaymentOptions, getPaymentRequests } from "src/adapters/api/payments";
import { notifyErrors } from "src/adapters/parsers";
import IconDots from "src/assets/icons/dots-vertical.svg?react";
import IconInfoCircle from "src/assets/icons/info-circle.svg?react";
import IconPaymentOptions from "src/assets/icons/payment-options.svg?react";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { TableCard } from "src/components/shared/TableCard";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { AppError, PaymentOption, PaymentRequest } from "src/domain";
import { ROUTES } from "src/routes";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { DETAILS, STATUS } from "src/utils/constants";
import { formatDate, formatIdentifier } from "src/utils/forms";

export function PaymentRequestsTable() {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();
  const navigate = useNavigate();

  const [paymentRequests, setPaymentRequests] = useState<AsyncTask<PaymentRequest[], AppError>>({
    status: "pending",
  });

  const [paymentOptions, setPaymentOptions] = useState<AsyncTask<PaymentOption[], AppError>>({
    status: "pending",
  });

  const paymentRequestsList = isAsyncTaskDataAvailable(paymentRequests) ? paymentRequests.data : [];

  const tableColumns: TableColumnsType<PaymentRequest> = isAsyncTaskDataAvailable(paymentOptions)
    ? [
        {
          dataIndex: "userDID",
          key: "userDID",
          render: (userDID: PaymentRequest["userDID"]) => (
            <Tooltip title={userDID}>
              <Typography.Text strong>{formatIdentifier(userDID, { short: true })}</Typography.Text>
            </Tooltip>
          ),
          title: "User DID",
        },
        {
          dataIndex: "createdAt",
          key: "createdAt",
          render: (createdAt: PaymentRequest["createdAt"]) => (
            <Typography.Text>{formatDate(createdAt)}</Typography.Text>
          ),
          sorter: ({ createdAt: a }, { createdAt: b }) => b.getTime() - a.getTime(),
          title: "Created date",
        },
        {
          dataIndex: "modifiedAt",
          key: "modifiedAt",
          render: (modifiedAt: PaymentRequest["modifiedAt"]) => (
            <Typography.Text>{formatDate(modifiedAt)}</Typography.Text>
          ),
          sorter: ({ modifiedAt: a }, { modifiedAt: b }) => b.getTime() - a.getTime(),
          title: "Modified date",
        },
        {
          dataIndex: "paymentOptionID",
          key: "paymentOptionID",
          render: (paymentOptionID: PaymentRequest["paymentOptionID"]) => {
            const paymentOptionName =
              isAsyncTaskDataAvailable(paymentOptions) &&
              paymentOptions.data.find(({ id }) => id === paymentOptionID)?.name;

            return (
              <Typography.Link
                href={generatePath(ROUTES.paymentOptionDetails.path, {
                  paymentOptionID,
                })}
              >
                {paymentOptionName || paymentOptionID}
              </Typography.Link>
            );
          },
          sorter: ({ createdAt: a }, { createdAt: b }) => b.getTime() - a.getTime(),
          title: "Payment option",
        },
        {
          dataIndex: "status",
          key: "status",
          render: (status: PaymentRequest["status"], { id }: PaymentRequest) => {
            return (
              <Row>
                {(() => {
                  switch (status) {
                    case PaymentRequestStatus.canceled:
                    case PaymentRequestStatus.failed: {
                      return <Tag color="error">{status}</Tag>;
                    }
                    case PaymentRequestStatus["not-verified"]:
                    case PaymentRequestStatus.pending: {
                      return <Tag>{status}</Tag>;
                    }
                    case PaymentRequestStatus.success: {
                      return <Tag color="success">{status}</Tag>;
                    }
                  }
                })()}
                <Dropdown
                  menu={{
                    items: [
                      {
                        icon: <IconInfoCircle />,
                        key: "details",
                        label: DETAILS,
                        onClick: () =>
                          navigate(
                            generatePath(ROUTES.paymentRequestDetils.path, {
                              paymentRequestID: id,
                            })
                          ),
                      },
                    ],
                  }}
                >
                  <Row>
                    <IconDots className="icon-secondary" />
                  </Row>
                </Dropdown>
              </Row>
            );
          },
          sorter: ({ status: a }, { status: b }) => a.localeCompare(b),
          title: STATUS,
        },
      ]
    : [];

  const fetchPaymentRequest = useCallback(
    async (signal?: AbortSignal) => {
      setPaymentRequests((previousState) =>
        isAsyncTaskDataAvailable(previousState)
          ? { data: previousState.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await getPaymentRequests({
        env,
        identifier,
        signal,
      });

      if (response.success) {
        if (response.data.failed.length) {
          void notifyErrors(response.data.failed);
        }

        setPaymentRequests({ data: response.data.successful, status: "successful" });
      } else {
        if (!isAbortedError(response.error)) {
          setPaymentRequests({ error: response.error, status: "failed" });
        }
      }
    },
    [env, identifier]
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

  return (
    <TableCard
      defaultContents={
        <>
          <Avatar className="avatar-color-icon" icon={<IconPaymentOptions />} size={48} />

          <Typography.Text strong>No payment requests</Typography.Text>
        </>
      }
      isLoading={isAsyncTaskStarting(paymentRequests) || isAsyncTaskStarting(paymentOptions)}
      showDefaultContents={
        paymentRequests.status === "successful" && paymentRequestsList.length === 0
      }
      table={
        <Table
          columns={tableColumns}
          dataSource={paymentRequestsList}
          locale={{
            emptyText: paymentRequests.status === "failed" && (
              <ErrorResult error={paymentRequests.error.message} />
            ),
          }}
          pagination={false}
          rowKey="id"
          showSorterTooltip
          sortDirections={["ascend", "descend"]}
          tableLayout="fixed"
        />
      }
      title={
        <Row justify="space-between">
          <Space size="middle">
            <Card.Meta title="Payment Requests" />
            <Tag>{paymentRequestsList.length}</Tag>
          </Space>
        </Row>
      }
    />
  );
}
