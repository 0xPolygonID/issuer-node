import { Avatar, Card, Divider, Row, Space, Table, Tag, Tooltip, Typography } from "antd";
import { ColumnsType } from "antd/es/table";
import { useCallback, useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";

import { APIError } from "src/adapters/api";
import { Connection, getConnections } from "src/adapters/api/connections";
import { Credential } from "src/adapters/api/credentials";
import { ReactComponent as IconUsers } from "src/assets/icons/users-01.svg";
import { ConnectionsRowDropdown } from "src/components/connections/ConnectionsRowDropdown";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { NoResults } from "src/components/shared/NoResults";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { TableCard } from "src/components/shared/TableCard";
import { useEnvContext } from "src/contexts/env";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { CONNECTIONS, IDENTIFIER, QUERY_SEARCH_PARAM } from "src/utils/constants";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/types";

export function Connections() {
  const env = useEnvContext();
  const [connections, setConnections] = useState<AsyncTask<Connection[], APIError>>({
    status: "pending",
  });

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
      render: (credentials: Credential[]) => (
        <Typography.Text>
          {[...credentials]
            .sort((a, b) => a.attributes.type.localeCompare(b.attributes.type))
            .map((credential) => credential.attributes.type)
            .join(", ")}
        </Typography.Text>
      ),
      title: "Issued credentials",
    },
    {
      dataIndex: "id",
      key: "id",
      render: (id: Connection["id"]) => <ConnectionsRowDropdown id={id} />,
      width: 55,
    },
  ];

  const fetchConnections = useCallback(
    async (signal: AbortSignal) => {
      setConnections({ status: "loading" });
      const response = await getConnections({
        credentials: true,
        env,
        params: {
          query: queryParam || undefined,
        },
        signal,
      });
      if (response.isSuccessful) {
        setConnections({ data: response.data, status: "successful" });
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
      setSearchParams((oldParams) => {
        const oldQuery = oldParams.get(QUERY_SEARCH_PARAM);
        const params = new URLSearchParams(oldParams);

        if (query === "") {
          params.delete(QUERY_SEARCH_PARAM);

          return params;
        } else if (oldQuery !== query) {
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
            <Space align="end" size="middle">
              <Card.Meta title={CONNECTIONS} />

              <Tag color="blue">{connectionsList.length}</Tag>
            </Space>
          </Row>
        }
      />
    </SiderLayoutContent>
  );
}
