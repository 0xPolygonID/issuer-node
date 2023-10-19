import { Avatar, Button, Card, Row, Space, Tag, Tooltip, Typography, message } from "antd";
import Table, { ColumnsType } from "antd/es/table";
import dayjs from "dayjs";
import { useCallback, useEffect, useState } from "react";
import { getNotification, markAsRead } from "src/adapters/api/notification";
//import { useSearchParams } from "react-router-dom";

import { ReactComponent as IconCreditCardRefresh } from "src/assets/icons/credit-card-refresh.svg";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { NoResults } from "src/components/shared/NoResults";
import { TableCard } from "src/components/shared/TableCard";
import { useEnvContext } from "src/contexts/Env";
import { AppError, Notification } from "src/domain";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { NOTIFICATION, REQUEST_DATE } from "src/utils/constants";
import { notifyParseErrors } from "src/utils/error";
import { formatDate } from "src/utils/forms";

export function NotificationsTable() {
  const env = useEnvContext();

  const [notifications, setNotifications] = useState<AsyncTask<Notification[], AppError>>({
    status: "pending",
  });
  const [messageAPI, messageContext] = message.useMessage();
  const User = localStorage.getItem("user");
  // const [searchParams, setSearchParams] = useSearchParams();

  // const statusParam = searchParams.get(STATUS_SEARCH_PARAM);
  const queryParam = ""; //searchParams.get(QUERY_SEARCH_PARAM);
  // const parsedStatusParam = requestStatusParser.safeParse(statusParam);
  //const notificationStatus = "all"; //parsedStatusParam.success ? parsedStatusParam.data : "all";

  const notificationsList = isAsyncTaskDataAvailable(notifications) ? notifications.data : [];
  const showDefaultContent =
    notifications.status === "successful" && notificationsList.length === 0 && queryParam === null;
  const deleteNotification = ({ Notification }: { Notification: Notification }) => {
    const notiId = Notification.id;
    void markAsRead({ env, notiId }).then((response) => {
      if (response) {
        void fetchNotifications();
        void messageAPI.success("Notification is marked as read");
      } else {
        void messageAPI.error("Something went wrong");
      }
    });
  };
  const tableColumns: ColumnsType<Notification> = [
    {
      dataIndex: "notification_message",
      key: "notification_message",
      render: (notification_message: Notification["notification_message"]) => (
        <Tooltip placement="topLeft" title={notification_message}>
          <Typography.Text strong>{notification_message}</Typography.Text>
        </Tooltip>
      ),
      title: "Message",
      width: "60%",
    },
    {
      dataIndex: "created_at",
      key: "created_at",
      render: (created_at: Notification["created_at"]) => (
        <Typography.Text>{formatDate(created_at)}</Typography.Text>
      ),
      sorter: ({ created_at: a }, { created_at: b }) => dayjs(a).unix() - dayjs(b).unix(),
      title: "Date",
      width: "20%",
    },
    {
      key: "operation",
      render: (Notification: Notification) => (
        <Button onClick={() => deleteNotification({ Notification })} type="primary">
          Mark as Read
        </Button>
      ),
      title: "Action",
      width: 100,
    },
  ];

  const fetchNotifications = useCallback(
    async (signal?: AbortSignal) => {
      setNotifications((previousNotifications) =>
        isAsyncTaskDataAvailable(previousNotifications)
          ? { data: previousNotifications.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await getNotification({
        env,
        module: User === "issuer" ? "Issuer" : User === "verifier" ? "Verifier" : "User",
        params: {
          query: queryParam || undefined,
        },
        signal,
      });

      if (response.success) {
        setNotifications({
          data: response.data.successful,
          status: "successful",
        });
        notifyParseErrors(response.data.failed);
      } else {
        if (!isAbortedError(response.error)) {
          setNotifications({ error: response.error, status: "failed" });
        }
      }
    },
    [env, queryParam, User]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchNotifications);

    return aborter;
  }, [fetchNotifications]);

  return (
    <>
      {messageContext}
      <TableCard
        defaultContents={
          <>
            <Avatar className="avatar-color-cyan" icon={<IconCreditCardRefresh />} size={48} />

            <Typography.Text strong>No Notifications</Typography.Text>

            <Typography.Text type="secondary">Notification will be listed here.</Typography.Text>
          </>
        }
        isLoading={isAsyncTaskStarting(notifications)}
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
            dataSource={notificationsList}
            locale={{
              emptyText:
                notifications.status === "failed" ? (
                  <ErrorResult error={notifications.error.message} />
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

              <Tag color="blue">{notificationsList.length}</Tag>
            </Space>
          </Row>
        }
      />
    </>
  );
}
