import { Avatar, Card, Row, Space, Tag, Tooltip, Typography } from "antd";
import Table, { ColumnsType } from "antd/es/table";
import dayjs from "dayjs";
import { useCallback, useEffect, useState } from "react";
//import { useSearchParams } from "react-router-dom";

import { getRequests } from "src/adapters/api/requests";
import { ReactComponent as IconCreditCardRefresh } from "src/assets/icons/credit-card-refresh.svg";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { NoResults } from "src/components/shared/NoResults";
import { TableCard } from "src/components/shared/TableCard";
import { useEnvContext } from "src/contexts/Env";
import { AppError } from "src/domain";
import { Request } from "src/domain/request";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { NOTIFICATION, REQUEST_DATE } from "src/utils/constants";
import { notifyParseErrors } from "src/utils/error";
import { formatDate } from "src/utils/forms";

export function NotificationsTable() {
  const env = useEnvContext();

  const [requests, setRequests] = useState<AsyncTask<Request[], AppError>>({
    status: "pending",
  });
  const User = localStorage.getItem("user");
  // const [searchParams, setSearchParams] = useSearchParams();

  // const statusParam = searchParams.get(STATUS_SEARCH_PARAM);
  const queryParam = ""; //searchParams.get(QUERY_SEARCH_PARAM);
  // const parsedStatusParam = requestStatusParser.safeParse(statusParam);
  const requestStatus = "all"; //parsedStatusParam.success ? parsedStatusParam.data : "all";

  const requestsList = isAsyncTaskDataAvailable(requests) ? requests.data : [];
  const showDefaultContent =
    requests.status === "successful" && requestsList.length === 0 && queryParam === null;

  let tableColumns: ColumnsType<Request>;
  if (User === "issuer") {
    tableColumns = [
      {
        dataIndex: "title",
        key: "title",
        render: (userDID: Request["userDID"]) => (
          <Tooltip placement="topLeft" title={userDID}>
            <Typography.Text strong>{userDID}</Typography.Text>
          </Tooltip>
        ),
        title: "Title",
        width: "20%",
      },
      {
        dataIndex: "message",
        key: "message",
        render: (credentialType: Request["credentialType"]) => (
          <Tooltip placement="topLeft" title={credentialType}>
            <Typography.Text strong>{credentialType}</Typography.Text>
          </Tooltip>
        ),
        title: "Message",
      },
      {
        dataIndex: "created_at",
        key: "requestDate",
        render: (requestDate: Request["requestDate"]) => (
          <Typography.Text>{formatDate(requestDate)}</Typography.Text>
        ),
        sorter: ({ requestDate: a }, { requestDate: b }) => dayjs(a).unix() - dayjs(b).unix(),
        title: REQUEST_DATE,
      },
    ];
  }

  const fetchRequests = useCallback(
    async (signal?: AbortSignal) => {
      setRequests((previousRequests) =>
        isAsyncTaskDataAvailable(previousRequests)
          ? { data: previousRequests.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await getRequests({
        env,
        params: {
          query: queryParam || undefined,
          status: requestStatus,
        },
        signal,
      });
      if (response.success) {
        setRequests({
          data: response.data.successful,
          status: "successful",
        });
        notifyParseErrors(response.data.failed);
      } else {
        if (!isAbortedError(response.error)) {
          setRequests({ error: response.error, status: "failed" });
        }
      }
    },
    [env, queryParam, requestStatus]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchRequests);

    return aborter;
  }, [fetchRequests]);

  return (
    <>
      <TableCard
        defaultContents={
          <>
            <Avatar className="avatar-color-cyan" icon={<IconCreditCardRefresh />} size={48} />

            <Typography.Text strong>No Notifications</Typography.Text>

            <Typography.Text type="secondary">Notification will be listed here.</Typography.Text>
          </>
        }
        isLoading={isAsyncTaskStarting(requests)}
        query={queryParam}
        searchPlaceholder="Search credentials, attributes, identifiers..."
        showDefaultContents={showDefaultContent}
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
            dataSource={requestsList}
            locale={{
              emptyText:
                requests.status === "failed" ? (
                  <ErrorResult error={requests.error.message} />
                ) : (
                  <NoResults searchQuery={queryParam} />
                ),
            }}
            pagination={false}
            rowKey="id"
            showSorterTooltip
            sortDirections={["ascend", "descend"]}
          />
        }
        title={
          <Row gutter={[0, 8]} justify="space-between">
            <Space size="middle">
              <Card.Meta title={NOTIFICATION} />

              <Tag color="blue">{requestsList.length}</Tag>
            </Space>
          </Row>
        }
      />
    </>
  );
}
