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

import { APIError } from "src/adapters/api";
import { getTransactions, publishState, retryPublishState } from "src/adapters/api/issuer-state";
import { ReactComponent as IconSwitch } from "src/assets/icons/switch-horizontal.svg";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { TableCard } from "src/components/shared/TableCard";
import { useEnvContext } from "src/contexts/Env";
import { useIssuerStateContext } from "src/contexts/IssuerState";
import { Transaction } from "src/domain";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";
import { makeRequestAbortable } from "src/utils/browser";

import { ISSUER_STATE, POLLING_INTERVAL, STATUS } from "src/utils/constants";
import { formatDate } from "src/utils/forms";

export function IssuerState() {
  const env = useEnvContext();
  const { refreshStatus, status } = useIssuerStateContext();

  const [isPublishing, setIsPublishing] = useState<boolean>(false);
  const [isRetrying, setIsRetrying] = useState<boolean>(false);
  const [transactions, setTransactions] = useState<AsyncTask<Transaction[], APIError>>({
    status: "pending",
  });

  const transactionsList = useMemo(
    () => (isAsyncTaskDataAvailable(transactions) ? transactions.data : []),
    [transactions]
  );

  const failedTransaction = transactionsList.find((transaction) => transaction.status === "failed");
  const disablePublishState =
    failedTransaction !== undefined || !isAsyncTaskDataAvailable(status) || !status.data;

  const publish = () => {
    setIsPublishing(true);
    void publishState({ env }).then((response) => {
      if (response.isSuccessful) {
        void message.success("Issuer state is being published");
      } else {
        void message.error(response.error.message);
      }

      void refreshStatus();
      void fetchTransactions();

      setIsPublishing(false);
    });
  };

  const retry = () => {
    setIsRetrying(true);
    void retryPublishState({ env }).then((response) => {
      if (response.isSuccessful) {
        void message.success("Issuer state is being published");
      } else {
        void message.error(response.error.message);
      }

      void refreshStatus();
      void fetchTransactions();

      setIsRetrying(false);
    });
  };

  const fetchTransactions = useCallback(
    async (signal?: AbortSignal) => {
      const response = await getTransactions({ env, signal });

      if (response.isSuccessful) {
        setTransactions({ data: response.data, status: "successful" });
      } else {
        setTransactions({ error: response.error, status: "failed" });
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
          <Link target="_blank" to={`${env.blockExplorerUrl}/tx/${txID}`}>
            {txID}
          </Link>
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
        <Typography.Text>{formatDate(publishDate, true)}</Typography.Text>
      ),
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
            return (
              <>
                <Tag color="error">{status}</Tag>
                <Button loading={isRetrying} onClick={retry} type="link">
                  Retry
                </Button>
              </>
            );
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
      description="Issuing merkle tree type credentials and revoking credentials require an additional step, know as publishing issuer state."
      title={ISSUER_STATE}
    >
      <Divider />
      <Space direction="vertical" size="large">
        <Card>
          <Row align="middle" justify="space-between">
            <Card.Meta
              title={disablePublishState ? "No pending actions" : "Pending actions to be published"}
            />
            <Button
              disabled={disablePublishState}
              loading={isPublishing || isAsyncTaskStarting(status)}
              onClick={publish}
              type="primary"
            >
              Publish issuer state
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
