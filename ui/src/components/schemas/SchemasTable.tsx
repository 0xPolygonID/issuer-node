import { Avatar, Button, Card, Row, Space, Table, Tag, Tooltip, Typography, message } from "antd";
import { ColumnsType } from "antd/es/table";
import { useCallback, useEffect, useState } from "react";
import { Link, generatePath, useSearchParams } from "react-router-dom";

import { Schema, schemasGetAll } from "src/adapters/api/schemas";
import { ReactComponent as IconSchema } from "src/assets/icons/file-search-02.svg";
import { ReactComponent as IconUpload } from "src/assets/icons/upload-01.svg";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { NoResults } from "src/components/shared/NoResults";
import { TableCard } from "src/components/shared/TableCard";
import { useEnvContext } from "src/contexts/env.context";
import { ROUTES } from "src/routes";
import { APIError, processZodError } from "src/utils/adapters";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { IMPORT_SCHEMA, QUERY_SEARCH_PARAM, SCHEMAS, SCHEMA_TYPE } from "src/utils/constants";
import { formatDate } from "src/utils/forms";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/types";

export function MySchemas() {
  const env = useEnvContext();
  const [schemas, setSchemas] = useState<AsyncTask<Schema[], APIError>>({
    status: "pending",
  });

  const [searchParams, setSearchParams] = useSearchParams();

  const queryParam = searchParams.get(QUERY_SEARCH_PARAM);

  const tableContents: ColumnsType<Schema> = [
    {
      dataIndex: "schema",
      ellipsis: { showTitle: false },
      key: "schema",
      render: (name: string) => (
        <Tooltip placement="topLeft" title={name}>
          <Typography.Text strong>{name}</Typography.Text>
        </Tooltip>
      ),
      sorter: ({ schema: a }, { schema: b }) => a.localeCompare(b),
      title: SCHEMA_TYPE,
    },
    {
      dataIndex: "createdAt",
      ellipsis: { showTitle: false },
      key: "createdAt",
      render: (createdAt: Date) => <Typography.Text>{formatDate(createdAt, true)}</Typography.Text>,
      sorter: ({ createdAt: a }, { createdAt: b }) => b.getTime() - a.getTime(),
      title: "Import date",
    },
    {
      dataIndex: "id",
      key: "id",
      render: (schemaID: string) => (
        <Link
          to={generatePath(ROUTES.issueCredential.path, {
            schemaID,
          })}
        >
          Issue
        </Link>
      ),
      title: "Actions",
    },
  ];

  const getSchemas = useCallback(
    async (signal: AbortSignal) => {
      setSchemas((oldState) =>
        isAsyncTaskDataAvailable(oldState)
          ? { data: oldState.data, status: "reloading" }
          : { status: "loading" }
      );
      const response = await schemasGetAll({
        env,
        params: {
          query: queryParam || undefined,
        },
        signal,
      });
      if (response.isSuccessful) {
        setSchemas({ data: response.data.schemas, status: "successful" });

        response.data.errors.forEach((zodError) => {
          processZodError(zodError).forEach((error) => void message.error(error));
        });
      } else {
        if (!isAbortedError(response.error)) {
          setSchemas({ error: response.error, status: "failed" });
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
    const { aborter } = makeRequestAbortable(getSchemas);

    return aborter;
  }, [getSchemas]);

  const schemaList = isAsyncTaskDataAvailable(schemas) ? schemas.data : [];

  return (
    <>
      <TableCard
        defaultContents={
          <>
            <Avatar className="avatar-color-cyan" icon={<IconSchema />} size={48} />

            <Typography.Text strong>No schemas</Typography.Text>

            <Typography.Text type="secondary">
              Imported schemas will be listed here.
            </Typography.Text>

            <Link to={ROUTES.importSchema.path}>
              <Button icon={<IconUpload />} type="primary">
                {IMPORT_SCHEMA}
              </Button>
            </Link>
          </>
        }
        isLoading={isAsyncTaskStarting(schemas)}
        onSearch={onSearch}
        query={queryParam}
        searchPlaceholder="Search schemas, attributes..."
        showDefaultContents={
          schemas.status === "successful" && schemaList.length === 0 && queryParam === null
        }
        table={
          <Table
            columns={tableContents.map(({ title, ...column }) => ({
              title: (
                <Typography.Text type="secondary">
                  <>{title}</>
                </Typography.Text>
              ),
              ...column,
            }))}
            dataSource={schemaList}
            locale={{
              emptyText:
                schemas.status === "failed" ? (
                  <ErrorResult error={schemas.error.message} />
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
              <Card.Meta title={SCHEMAS} />

              <Tag color="blue">{schemaList.length}</Tag>
            </Space>
          </Row>
        }
      />
    </>
  );
}
