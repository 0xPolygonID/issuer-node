import { Avatar, Card, Divider, Dropdown, Row, Space, Table, Tag, Tooltip, Typography } from "antd";
import { ColumnsType } from "antd/es/table";
import { useCallback, useEffect, useState } from "react";
import { generatePath, useNavigate, useSearchParams } from "react-router-dom";

import { getConnections } from "src/adapters/api/connections";
import { ReactComponent as IconCreditCardPlus } from "src/assets/icons/credit-card-plus.svg";
import { ReactComponent as IconDots } from "src/assets/icons/dots-vertical.svg";
import { ReactComponent as IconInfoCircle } from "src/assets/icons/info-circle.svg";
import { ReactComponent as IconTrash } from "src/assets/icons/trash-01.svg";
import { ReactComponent as IconUsers } from "src/assets/icons/users-01.svg";
import { ConnectionDeleteModal } from "src/components/connections/ConnectionDeleteModal";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { NoResults } from "src/components/shared/NoResults";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { TableCard } from "src/components/shared/TableCard";
import { useEnvContext } from "src/contexts/Env";
import { AppError, Connection } from "src/domain";
import { ROUTES } from "src/routes";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import {
  CONNECTIONS,
  DELETE,
  DETAILS,
  DID_SEARCH_PARAM,
  IDENTIFIER,
  ISSUED_CREDENTIALS,
  QUERY_SEARCH_PARAM,
} from "src/utils/constants";
import { notifyParseErrors } from "src/utils/error";

export function ConnectionsTable() {
  const env = useEnvContext();

  const navigate = useNavigate();

  const [connections, setConnections] = useState<AsyncTask<Connection[], AppError>>({
    status: "pending",
  });
  const [connectionToDelete, setConnectionToDelete] = useState<string>();

  const [searchParams, setSearchParams] = useSearchParams();

  const queryParam = searchParams.get(QUERY_SEARCH_PARAM);

  const tableColumns: ColumnsType<Connection> = [
    {
      dataIndex: "userID",
      ellipsis: { showTitle: false },
      key: "type",
      render: (userID: Connection["userID"]) => (
        <Tooltip placement="topLeft" title={userID}>
          <Typography.Text strong>{userID.split(":").pop()}</Typography.Text>
        </Tooltip>
      ),
      sorter: ({ id: a }, { id: b }) => a.localeCompare(b),
      title: IDENTIFIER,
    },
    {
      dataIndex: "credentials",
      ellipsis: { showTitle: false },
      key: "credentials",
      render: (credentials: Connection["credentials"]) => (
        <Typography.Text>
          {[...credentials.successful]
            .sort((a, b) => a.schemaType.localeCompare(b.schemaType))
            .map((credential) => credential.schemaType)
            .join(", ")}
        </Typography.Text>
      ),
      title: ISSUED_CREDENTIALS,
    },
    {
      dataIndex: "id",
      key: "id",
      render: (id: Connection["id"], { userID }: Connection) => (
        <Dropdown
          menu={{
            items: [
              {
                icon: <IconInfoCircle />,
                key: "details",
                label: DETAILS,
                onClick: () =>
                  navigate(generatePath(ROUTES.connectionDetails.path, { connectionID: id })),
              },
              {
                key: "divider1",
                type: "divider",
              },
              {
                icon: <IconCreditCardPlus />,
                key: "issue",
                label: "Issue credential directly",
                onClick: () =>
                  navigate({
                    pathname: generatePath(ROUTES.issueCredential.path),
                    search: `${DID_SEARCH_PARAM}=${userID}`,
                  }),
              },
              {
                key: "divider2",
                type: "divider",
              },
              {
                danger: true,
                icon: <IconTrash />,
                key: "delete",
                label: DELETE,
                onClick: () => setConnectionToDelete(id),
              },
            ],
          }}
          overlayStyle={{ zIndex: 999 }}
        >
          <Row>
            <IconDots className="icon-secondary" />
          </Row>
        </Dropdown>
      ),
      width: 55,
    },
  ];

  const fetchConnections = useCallback(
    async (signal?: AbortSignal) => {
      setConnections({ status: "loading" });
      const response = await getConnections({
        credentials: true,
        env,
        params: {
          query: queryParam || undefined,
        },
        signal,
      });
      if (response.success) {
        setConnections({ data: response.data.successful, status: "successful" });
        notifyParseErrors(response.data.failed);
      } else {
        if (!isAbortedError(response.error)) {
          setConnections({ error: response.error, status: "failed" });
        }
      }
    },
    [env, queryParam]
  );

  const onSearch = useCallback(
    (query: string) => {
      setSearchParams((previousParams) => {
        const previousQuery = previousParams.get(QUERY_SEARCH_PARAM);
        const params = new URLSearchParams(previousParams);

        if (query === "") {
          params.delete(QUERY_SEARCH_PARAM);

          return params;
        } else if (previousQuery !== query) {
          params.set(QUERY_SEARCH_PARAM, query);

          return params;
        }
        return params;
      });
    },
    [setSearchParams]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchConnections);

    return aborter;
  }, [fetchConnections]);

  const connectionsList = isAsyncTaskDataAvailable(connections) ? connections.data : [];

  return (
    <SiderLayoutContent
      description="Connections are established via a secure channel upon issuing credentials to users."
      title={CONNECTIONS}
    >
      <Divider />

      <TableCard
        defaultContents={
          <>
            <Avatar className="avatar-color-cyan" icon={<IconUsers />} size={48} />

            <Typography.Text strong>No connections</Typography.Text>

            <Typography.Text type="secondary">
              Your connections will be listed here.
            </Typography.Text>
          </>
        }
        isLoading={isAsyncTaskStarting(connections)}
        onSearch={onSearch}
        query={queryParam}
        searchPlaceholder="Search connections, credentials..."
        showDefaultContents={
          connections.status === "successful" && connectionsList.length === 0 && queryParam === null
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
            dataSource={connectionsList}
            locale={{
              emptyText:
                connections.status === "failed" ? (
                  <ErrorResult error={connections.error.message} />
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
          <Row justify="space-between">
            <Space size="middle">
              <Card.Meta title={CONNECTIONS} />

              <Tag color="blue">{connectionsList.length}</Tag>
            </Space>
          </Row>
        }
      />
      {connectionToDelete && (
        <ConnectionDeleteModal
          id={connectionToDelete}
          onClose={() => setConnectionToDelete(undefined)}
          onDelete={() => void fetchConnections()}
        />
      )}
    </SiderLayoutContent>
  );
}
