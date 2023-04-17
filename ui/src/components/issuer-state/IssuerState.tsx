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
import { useCallback, useEffect, useState } from "react";

import { APIError } from "src/adapters/api";
import { getTransactions, publishState } from "src/adapters/api/issuer-state";
import { ReactComponent as IconSwitch } from "src/assets/icons/switch-horizontal.svg";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { TableCard } from "src/components/shared/TableCard";
import { useEnvContext } from "src/contexts/env";
import { useStateContext } from "src/contexts/issuer-state";
import { Transaction } from "src/domain";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";
import { makeRequestAbortable } from "src/utils/browser";

import { ISSUER_STATE, STATUS } from "src/utils/constants";
import { formatDate } from "src/utils/forms";

export function IssuerState() {
  const env = useEnvContext();
  const { refreshStatus, status } = useStateContext();
  const [transactions, setTransactions] = useState<AsyncTask<Transaction[], APIError>>({
    status: "pending",
  });

  const transactionsList = isAsyncTaskDataAvailable(transactions) ? transactions.data : [];

  const publish = () => {
    void publishState({ env }).then((response) => {
      if (response.isSuccessful) {
        void message.success("Issuer state is being published");
      } else {
        void message.error(response.error.message);
      }

      void refreshStatus();
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
        <Tooltip placement="topLeft" title={txID}>
          <Typography.Text strong>{txID}</Typography.Text>
        </Tooltip>
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
          case "published": {
            return <Tag color="success">{status}</Tag>;
          }
          case "failed": {
            return (
              <>
                <Tag color="error">{status}</Tag>
                <Button>Retry</Button>
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

  return (
    <SiderLayoutContent
      description="Issuing merkle tree type credentials and revoking credentials require an additional step, know as publishing issuer state."
      title={ISSUER_STATE}
    >
      <Divider />
      <Space direction="vertical" size="large">
        <Card>
          <Row align="middle" justify="space-between">
            <Card.Meta title={status ? "Pending actions to be published" : "No pending actions"} />
            <Button
              disabled={!isAsyncTaskDataAvailable(status) || !status.data}
              loading={isAsyncTaskStarting(status)}
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
