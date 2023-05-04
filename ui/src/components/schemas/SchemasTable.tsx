import { Avatar, Button, Card, Row, Space, Table, Tag, Tooltip, Typography } from "antd";
import { ColumnsType } from "antd/es/table";
import { useCallback, useEffect, useState } from "react";
import { Link, generatePath, useSearchParams } from "react-router-dom";

import { getSchemas } from "src/adapters/api/schemas";
import { ReactComponent as IconSchema } from "src/assets/icons/file-search-02.svg";
import { ReactComponent as IconUpload } from "src/assets/icons/upload-01.svg";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { NoResults } from "src/components/shared/NoResults";
import { TableCard } from "src/components/shared/TableCard";
import { useEnvContext } from "src/contexts/Env";
import { AppError, Schema } from "src/domain";
import { ROUTES } from "src/routes";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import {
  IMPORT_SCHEMA,
  QUERY_SEARCH_PARAM,
  SCHEMAS,
  SCHEMA_SEARCH_PARAM,
  SCHEMA_TYPE,
} from "src/utils/constants";
import { notifyParseErrors } from "src/utils/error";
import { formatDate } from "src/utils/forms";

export function MySchemas() {
  const env = useEnvContext();
  const [schemas, setSchemas] = useState<AsyncTask<Schema[], AppError>>({
    status: "pending",
  });

  const [searchParams, setSearchParams] = useSearchParams();

  const queryParam = searchParams.get(QUERY_SEARCH_PARAM);

  const tableColumns: ColumnsType<Schema> = [
    {
      dataIndex: "type",
      ellipsis: { showTitle: false },
      key: "type",
      render: (type: Schema["type"]) => (
        <Tooltip placement="topLeft" title={type}>
          <Typography.Text strong>{type}</Typography.Text>
        </Tooltip>
      ),
      sorter: ({ type: a }, { type: b }) => a.localeCompare(b),
      title: SCHEMA_TYPE,
    },
    {
      dataIndex: "createdAt",
      ellipsis: { showTitle: false },
      key: "createdAt",
      render: (createdAt: Schema["createdAt"]) => (
        <Typography.Text>{formatDate(createdAt)}</Typography.Text>
      ),
      sorter: ({ createdAt: a }, { createdAt: b }) => b.getTime() - a.getTime(),
      title: "Import date",
    },
    {
      dataIndex: "id",
      key: "id",
      render: (schemaID: Schema["id"]) => (
        <Row>
          <Space size="large">
            <Link
              to={generatePath(ROUTES.schemaDetails.path, {
                schemaID,
              })}
            >
              Details
            </Link>
            <Link
              to={{
                pathname: generatePath(ROUTES.issueCredential.path),
                search: `${SCHEMA_SEARCH_PARAM}=${schemaID}`,
              }}
            >
              Issue
            </Link>
          </Space>
        </Row>
      ),
      title: "Actions",
    },
  ];

  const onGetSchemas = useCallback(
    async (signal: AbortSignal) => {
      setSchemas((previousState) =>
        isAsyncTaskDataAvailable(previousState)
          ? { data: previousState.data, status: "reloading" }
          : { status: "loading" }
      );
      const response = await getSchemas({
        env,
        params: {
          query: queryParam || undefined,
        },
        signal,
      });
      if (response.success) {
        setSchemas({ data: response.data.successful, status: "successful" });
        notifyParseErrors(response.data.failed);
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
    const { aborter } = makeRequestAbortable(onGetSchemas);

    return aborter;
  }, [onGetSchemas]);

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
            columns={tableColumns.map(({ title, ...column }) => ({
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
