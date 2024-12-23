import {
  App,
  Avatar,
  Button,
  Card,
  Divider,
  Row,
  Space,
  Table,
  TableColumnsType,
  Tag,
  Tooltip,
  Typography,
} from "antd";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { Sorter, parseSorters, serializeSorters } from "src/adapters/api";

import { getTransactions, publishState, retryPublishState } from "src/adapters/api/issuer-state";
import { notifyErrors, positiveIntegerFromStringParser } from "src/adapters/parsers";
import { tableSorterParser } from "src/adapters/parsers/view";
import IconAlert from "src/assets/icons/alert-circle.svg?react";
import IconSwitch from "src/assets/icons/switch-horizontal.svg?react";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { TableCard } from "src/components/shared/TableCard";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { useIssuerStateContext } from "src/contexts/IssuerState";
import { AppError, Transaction } from "src/domain";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";

import {
  DEFAULT_PAGINATION_MAX_RESULTS,
  DEFAULT_PAGINATION_PAGE,
  DEFAULT_PAGINATION_TOTAL,
  ISSUER_STATE,
  PAGINATION_MAX_RESULTS_PARAM,
  PAGINATION_PAGE_PARAM,
  POLLING_INTERVAL,
  QUERY_SEARCH_PARAM,
  SORT_PARAM,
  STATUS,
} from "src/utils/constants";
import { formatDate } from "src/utils/forms";

const PUBLISHED_MESSAGE = "Issuer state is being published";

export function IssuerState() {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();
  const { refreshStatus, status } = useIssuerStateContext();

  const [isPublishing, setIsPublishing] = useState<boolean>(false);
  const [transactions, setTransactions] = useState<AsyncTask<Transaction[], AppError>>({
    status: "pending",
  });

  const { message } = App.useApp();
  const transactionsList = useMemo(
    () => (isAsyncTaskDataAvailable(transactions) ? transactions.data : []),
    [transactions]
  );

  const [searchParams, setSearchParams] = useSearchParams();

  const queryParam = searchParams.get(QUERY_SEARCH_PARAM);
  const paginationPageParam = searchParams.get(PAGINATION_PAGE_PARAM);
  const paginationMaxResultsParam = searchParams.get(PAGINATION_MAX_RESULTS_PARAM);
  const sortParam = searchParams.get(SORT_PARAM);

  const sorters = parseSorters(sortParam);
  const paginationPageParsed = positiveIntegerFromStringParser.safeParse(paginationPageParam);
  const paginationMaxResultsParsed =
    positiveIntegerFromStringParser.safeParse(paginationMaxResultsParam);

  const [paginationTotal, setPaginationTotal] = useState<number>(DEFAULT_PAGINATION_TOTAL);

  const paginationPage = paginationPageParsed.success
    ? paginationPageParsed.data
    : DEFAULT_PAGINATION_PAGE;
  const paginationMaxResults = paginationMaxResultsParsed.success
    ? paginationMaxResultsParsed.data
    : DEFAULT_PAGINATION_MAX_RESULTS;

  const failedTransaction = transactionsList.find((transaction) => transaction.status === "failed");
  const disablePublishState = !isAsyncTaskDataAvailable(status) || !status.data;

  const publish = () => {
    setIsPublishing(true);

    const functionToExecute = failedTransaction ? retryPublishState : publishState;

    void functionToExecute({ env, identifier }).then((response) => {
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

  const updateUrlParams = useCallback(
    ({ maxResults, page, sorters }: { maxResults?: number; page?: number; sorters?: Sorter[] }) => {
      setSearchParams((previousParams) => {
        const params = new URLSearchParams(previousParams);
        params.set(
          PAGINATION_PAGE_PARAM,
          page !== undefined ? page.toString() : DEFAULT_PAGINATION_PAGE.toString()
        );
        params.set(
          PAGINATION_MAX_RESULTS_PARAM,
          maxResults !== undefined
            ? maxResults.toString()
            : DEFAULT_PAGINATION_MAX_RESULTS.toString()
        );
        const newSorters = sorters || parseSorters(sortParam);
        newSorters.length > 0
          ? params.set(SORT_PARAM, serializeSorters(newSorters))
          : params.delete(SORT_PARAM);

        return params;
      });
    },
    [setSearchParams, sortParam]
  );

  const fetchTransactions = useCallback(
    async (signal?: AbortSignal) => {
      setTransactions((previousTransactions) =>
        isAsyncTaskDataAvailable(previousTransactions)
          ? { data: previousTransactions.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await getTransactions({
        env,
        identifier,
        params: {
          maxResults: paginationMaxResults,
          page: paginationPage,
          query: queryParam || undefined,
          sorters: parseSorters(sortParam),
        },
        signal,
      });

      if (response.success) {
        setTransactions({ data: response.data.items.successful, status: "successful" });
        setPaginationTotal(response.data.meta.total);
        updateUrlParams({
          maxResults: response.data.meta.max_results,
          page: response.data.meta.page,
        });
        void notifyErrors(response.data.items.failed);
      } else {
        if (!isAbortedError(response.error)) {
          setTransactions({ error: response.error, status: "failed" });
        }
      }
    },
    [env, identifier, paginationMaxResults, paginationPage, queryParam, sortParam, updateUrlParams]
  );

  const tableColumns: TableColumnsType<Transaction> = [
    {
      dataIndex: "txID",
      ellipsis: { showTitle: true },
      key: "txID",
      render: (txID: Transaction["txID"]) => (
        <>
          {txID ? (
            <Tooltip placement="topLeft" title={txID}>
              <Typography.Text>{txID ? txID : "-"}</Typography.Text>
            </Tooltip>
          ) : (
            <Typography.Text>-</Typography.Text>
          )}
        </>
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
      sorter: true,
      sortOrder: sorters?.find(({ field }) => field === "publishDate")?.order,
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
      sorter: true,
      sortOrder: sorters?.find(({ field }) => field === "status")?.order,
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
              <Avatar className="avatar-color-icon" icon={<IconSwitch />} size={48} />

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
              onChange={({ current, pageSize, total }, _, sorters) => {
                setPaginationTotal(total || DEFAULT_PAGINATION_TOTAL);
                const parsedSorters = tableSorterParser.safeParse(sorters);
                updateUrlParams({
                  maxResults: pageSize,
                  page: current,
                  sorters: parsedSorters.success ? parsedSorters.data : [],
                });
              }}
              pagination={{
                current: paginationPage,
                hideOnSinglePage: true,
                pageSize: paginationMaxResults,
                position: ["bottomRight"],
                total: paginationTotal,
              }}
              rowKey="id"
              showSorterTooltip
              sortDirections={["ascend", "descend"]}
            />
          }
          title={
            <Row justify="space-between">
              <Space align="end" size="middle">
                <Card.Meta title="Published states" />

                <Tag>{transactionsList.length}</Tag>
              </Space>
            </Row>
          }
        />
      </Space>
    </SiderLayoutContent>
  );
}
