import {
  Avatar,
  Button,
  Card,
  Grid,
  Row,
  Space,
  Table,
  TableColumnsType,
  Tag,
  Tooltip,
  Typography,
} from "antd";

import { useCallback, useEffect, useState } from "react";
import { Link, generatePath, useNavigate, useSearchParams } from "react-router-dom";

import { getApiSchemas } from "src/adapters/api/schemas";
import { notifyErrors } from "src/adapters/parsers";
import IconSchema from "src/assets/icons/file-search-02.svg?react";
import IconUpload from "src/assets/icons/upload-01.svg?react";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { NoResults } from "src/components/shared/NoResults";
import { TableCard } from "src/components/shared/TableCard";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { ApiSchema, AppError } from "src/domain";
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
import { formatDate } from "src/utils/forms";

export function SchemasTable() {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();
  const navigate = useNavigate();
  const [apiSchemas, setApiSchemas] = useState<AsyncTask<ApiSchema[], AppError>>({
    status: "pending",
  });

  const { sm } = Grid.useBreakpoint();

  const [searchParams, setSearchParams] = useSearchParams();

  const queryParam = searchParams.get(QUERY_SEARCH_PARAM);

  const tableColumns: TableColumnsType<ApiSchema> = [
    {
      dataIndex: "type",
      ellipsis: { showTitle: false },
      key: "type",
      render: (type: ApiSchema["type"], { description, id, title }: ApiSchema) => (
        <Tooltip
          placement="topLeft"
          title={title && description ? `${title}: ${description}` : title || description}
        >
          <Typography.Link
            onClick={() =>
              navigate(
                generatePath(ROUTES.schemaDetails.path, {
                  schemaID: id,
                })
              )
            }
            strong
          >
            {type}
          </Typography.Link>
        </Tooltip>
      ),
      sorter: {
        compare: ({ type: a }, { type: b }) => a.localeCompare(b),
        multiple: 2,
      },
      title: SCHEMA_TYPE,
    },
    {
      dataIndex: "version",
      key: "version",
      render: (version: ApiSchema["version"]) => (
        <Typography.Text>{version || "-"}</Typography.Text>
      ),
      sorter: {
        compare: ({ version: a }, { version: b }) => (a && b ? a.localeCompare(b) : 0),
        multiple: 1,
      },
      title: "Schema version",
    },
    {
      dataIndex: "createdAt",
      key: "createdAt",
      render: (createdAt: ApiSchema["createdAt"]) => (
        <Typography.Text>{formatDate(createdAt)}</Typography.Text>
      ),
      sorter: ({ createdAt: a }, { createdAt: b }) => b.getTime() - a.getTime(),
      title: "Import date",
    },
    {
      dataIndex: "id",
      key: "id",
      render: (schemaID: ApiSchema["id"]) => (
        <Row>
          <Space size="large">
            <Link
              to={generatePath(ROUTES.schemaDetails.path, {
                schemaID,
              })}
            >
              Details
            </Link>
            {sm && (
              <Link
                to={{
                  pathname: generatePath(ROUTES.issueCredential.path),
                  search: `${SCHEMA_SEARCH_PARAM}=${schemaID}`,
                }}
              >
                Issue
              </Link>
            )}
          </Space>
        </Row>
      ),
      title: "Actions",
    },
  ];

  const onGetSchemas = useCallback(
    async (signal: AbortSignal) => {
      setApiSchemas((previousState) =>
        isAsyncTaskDataAvailable(previousState)
          ? { data: previousState.data, status: "reloading" }
          : { status: "loading" }
      );
      const response = await getApiSchemas({
        env,
        identifier,
        params: {
          query: queryParam || undefined,
        },
        signal,
      });
      if (response.success) {
        setApiSchemas({ data: response.data.successful, status: "successful" });
        void notifyErrors(response.data.failed);
      } else {
        if (!isAbortedError(response.error)) {
          setApiSchemas({ error: response.error, status: "failed" });
        }
      }
    },
    [env, queryParam, identifier]
  );

  const onSearch = useCallback(
    (query: string) => {
      setSearchParams((previousParams) => {
        const previousQuery = previousParams.get(QUERY_SEARCH_PARAM);
        const params = new URLSearchParams(previousParams);

        if (query === "") {
          params.delete(QUERY_SEARCH_PARAM);
        } else if (previousQuery !== query) {
          params.set(QUERY_SEARCH_PARAM, query);
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

  const schemaList = isAsyncTaskDataAvailable(apiSchemas) ? apiSchemas.data : [];

  return (
    <TableCard
      defaultContents={
        <>
          <Avatar className="avatar-color-icon" icon={<IconSchema />} size={48} />

          <Typography.Text strong>No schemas</Typography.Text>

          <Typography.Text type="secondary">Imported schemas will be listed here.</Typography.Text>

          <Link to={ROUTES.importSchema.path}>
            <Button icon={<IconUpload />} type="primary">
              {IMPORT_SCHEMA}
            </Button>
          </Link>
        </>
      }
      isLoading={isAsyncTaskStarting(apiSchemas)}
      onSearch={onSearch}
      query={queryParam}
      searchPlaceholder="Search schemas, attributes..."
      showDefaultContents={
        apiSchemas.status === "successful" && schemaList.length === 0 && queryParam === null
      }
      table={
        <Table
          columns={tableColumns}
          dataSource={schemaList}
          locale={{
            emptyText:
              apiSchemas.status === "failed" ? (
                <ErrorResult error={apiSchemas.error.message} />
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
            <Card.Meta title={SCHEMAS} />

            <Tag>{schemaList.length}</Tag>
          </Space>
        </Row>
      }
    />
  );
}
