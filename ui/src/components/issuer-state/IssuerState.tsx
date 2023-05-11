import {
  Avatar,
  Button,
  Card,
  Divider,
  Row,
  Space,
  Table,
  Tag,
  Tooltip,
  Typography,
  message,
} from "antd";
import { ColumnsType } from "antd/es/table";
import dayjs from "dayjs";
import { useCallback, useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";

import { getTransactions, publishState, retryPublishState } from "src/adapters/api/issuer-state";
import { ReactComponent as IconAlert } from "src/assets/icons/alert-circle.svg";
import { ReactComponent as IconSwitch } from "src/assets/icons/switch-horizontal.svg";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { TableCard } from "src/components/shared/TableCard";
import { useEnvContext } from "src/contexts/Env";
import { useIssuerStateContext } from "src/contexts/IssuerState";
import { AppError, Transaction } from "src/domain";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";

import { ISSUER_STATE, POLLING_INTERVAL, STATUS } from "src/utils/constants";
import { notifyParseErrors } from "src/utils/error";
import { formatDate } from "src/utils/forms";

const PUBLISHED_MESSAGE = "Issuer state is being published";

export function IssuerState() {
  const env = useEnvContext();
  const { refreshStatus, status } = useIssuerStateContext();

  const [isPublishing, setIsPublishing] = useState<boolean>(false);
  const [transactions, setTransactions] = useState<AsyncTask<Transaction[], AppError>>({
    status: "pending",
  });

  const transactionsList = useMemo(
    () => (isAsyncTaskDataAvailable(transactions) ? transactions.data : []),
    [transactions]
  );

  const failedTransaction = transactionsList.find((transaction) => transaction.status === "failed");
  const disablePublishState = !isAsyncTaskDataAvailable(status) || !status.data;

  const publish = () => {
    setIsPublishing(true);

    const functionToExecute = failedTransaction ? retryPublishState : publishState;

    void functionToExecute({ env }).then((response) => {
      if (response.success) {
        void message.success(PUBLISHED_MESSAGE);
      } else {
        void message.error(response.error.message);
      }

      void refreshStatus();
      void fetchTransactions();

      setIsPublishing(false);
    });
  };

  const fetchTransactions = useCallback(
    async (signal?: AbortSignal) => {
      const response = await getTransactions({ env, signal });

      if (response.success) {
        setTransactions({ data: response.data.successful, status: "successful" });
        notifyParseErrors(response.data.failed);
      } else {
        if (!isAbortedError(response.error)) {
          setTransactions({ error: response.error, status: "failed" });
        }
      }
    },
    [env]
  );

  const tableColumns: ColumnsType<Transaction> = [
    {
      dataIndex: "txID",
      ellipsis: { showTitle: false },
      key: "txID",
      render: (txID: Transaction["txID"]) => (
        <Typography.Text strong>
          {txID ? (
            <Link target="_blank" to={`${env.blockExplorerUrl}/tx/${txID}`}>
              {txID}
            </Link>
          ) : (
            "-"
          )}
        </Typography.Text>
      ),
      title: "Transaction ID",
    },
    {
      dataIndex: "state",
      ellipsis: { showTitle: false },
      key: "state",
      render: (state: Transaction["state"]) => (
        <Tooltip placement="topLeft" title={state}>
          <Typography.Text>{state}</Typography.Text>
        </Tooltip>
      ),
      title: "Identity state",
    },
    {
      dataIndex: "publishDate",
      ellipsis: { showTitle: false },
      key: "publishDate",
      render: (publishDate: Transaction["publishDate"]) => (
        <Typography.Text>{formatDate(publishDate)}</Typography.Text>
      ),
      responsive: ["sm"],
      sorter: ({ publishDate: a }, { publishDate: b }) => dayjs(a).unix() - dayjs(b).unix(),
      title: "Publish date",
    },
    {
      dataIndex: "status",
      ellipsis: { showTitle: false },
      key: "status",
      render: (status: Transaction["status"]) => {
        switch (status) {
          case "created":
          case "pending": {
            return <Tag color="warning">{status}</Tag>;
          }
          case "transacted":
          case "published": {
            return <Tag color="success">{status}</Tag>;
          }
          case "failed": {
            return <Tag color="error">{status}</Tag>;
          }
        }
      },
      sorter: ({ status: a }, { status: b }) => a.localeCompare(b),
      title: STATUS,
    },
  ];

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchTransactions);

    return aborter;
  }, [fetchTransactions]);

  useEffect(() => {
    const { aborter } = makeRequestAbortable(refreshStatus);

    return aborter;
  }, [refreshStatus]);

  useEffect(() => {
    const checkUnpublished = setInterval(() => {
      const isUnpublished = transactionsList.find(
        (transaction) => transaction.status !== "published"
      );

      if (isUnpublished) {
        void fetchTransactions();
      } else {
        clearInterval(checkUnpublished);
      }
    }, POLLING_INTERVAL);

    return () => clearInterval(checkUnpublished);
  }, [fetchTransactions, transactionsList]);

  return (
    <SiderLayoutContent
      description="Issuing Merkle tree type credentials and revoking credentials require an additional step, known as publishing issuer state."
      title={ISSUER_STATE}
    >
      <Divider />

      <Space direction="vertical" size="large">
        <Card>
          <Row gutter={[0, 16]} justify="space-between">
            {disablePublishState ? (
              <Card.Meta title="No pending actions" />
            ) : (
              <Card.Meta
                avatar={
                  failedTransaction && (
                    <Avatar className="avatar-color-error" icon={<IconAlert />} />
                  )
                }
                description={
                  failedTransaction
                    ? "Please try again."
                    : "You can publish issuer state now or bulk publish with other actions."
                }
                title={
                  failedTransaction
                    ? "Transaction failed to publish"
                    : "Pending actions to be published"
                }
              />
            )}
            <Button
              disabled={disablePublishState}
              loading={isPublishing || isAsyncTaskStarting(status)}
              onClick={publish}
              type="primary"
            >
              {failedTransaction ? "Republish" : "Publish"} issuer state
            </Button>
          </Row>
        </Card>

        <TableCard
          defaultContents={
            <>
              <Avatar className="avatar-color-cyan" icon={<IconSwitch />} size={48} />

              <Typography.Text strong>No transactions</Typography.Text>

              <Typography.Text type="secondary">
                Published transactions will be listed here
              </Typography.Text>
            </>
          }
          isLoading={isAsyncTaskStarting(transactions)}
          showDefaultContents={
            transactions.status === "successful" && transactionsList.length === 0
          }
          table={
            <Table
              columns={tableColumns.map(({ title, ...column }) => ({
                title: (
                  <Typography.Text type="secondary">
                    <>{title}</>
                  </Typography.Text>
                ),
                ...column,
              }))}
              dataSource={transactionsList}
              locale={{
                emptyText: transactions.status === "failed" && (
                  <ErrorResult error={transactions.error.message} />
                ),
              }}
              pagination={false}
              rowKey="id"
              showSorterTooltip
              sortDirections={["ascend", "descend"]}
            />
          }
          title={
            <Row justify="space-between">
              <Space align="end" size="middle">
                <Card.Meta title="Published states" />

                <Tag color="blue">{transactionsList.length}</Tag>
              </Space>
            </Row>
          }
        />
      </Space>
    </SiderLayoutContent>
  );
}
