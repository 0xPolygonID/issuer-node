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
import { Link, generatePath, useSearchParams } from "react-router-dom";
import { Sorter, parseSorters, serializeSorters } from "src/adapters/api";

import { getApiSchemas } from "src/adapters/api/schemas";
import { positiveIntegerFromStringParser } from "src/adapters/parsers";
import { tableSorterParser } from "src/adapters/parsers/view";
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
  DEFAULT_PAGINATION_MAX_RESULTS,
  DEFAULT_PAGINATION_PAGE,
  DEFAULT_PAGINATION_TOTAL,
  IMPORT_SCHEMA,
  PAGINATION_MAX_RESULTS_PARAM,
  PAGINATION_PAGE_PARAM,
  QUERY_SEARCH_PARAM,
  SCHEMAS,
  SCHEMA_SEARCH_PARAM,
  SCHEMA_TYPE,
  SORT_PARAM,
} from "src/utils/constants";
import { notifyParseErrors } from "src/utils/error";
import { formatDate } from "src/utils/forms";

export function SchemasTable() {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();

  const [apiSchemas, setApiSchemas] = useState<AsyncTask<ApiSchema[], AppError>>({
    status: "pending",
  });
  const [paginationTotal, setPaginationTotal] = useState<number>(DEFAULT_PAGINATION_TOTAL);

  const { sm } = Grid.useBreakpoint();

  const [searchParams, setSearchParams] = useSearchParams();

  const queryParam = searchParams.get(QUERY_SEARCH_PARAM);
  const paginationPageParam = searchParams.get(PAGINATION_PAGE_PARAM);
  const paginationMaxResultsParam = searchParams.get(PAGINATION_MAX_RESULTS_PARAM);
  const sortParam = searchParams.get(SORT_PARAM);

  const sorters = parseSorters(sortParam);
  const paginationPageParsed = positiveIntegerFromStringParser.safeParse(paginationPageParam);
  const paginationMaxResultsParsed =
    positiveIntegerFromStringParser.safeParse(paginationMaxResultsParam);

  const paginationPage = paginationPageParsed.success
    ? paginationPageParsed.data
    : DEFAULT_PAGINATION_PAGE;
  const paginationMaxResults = paginationMaxResultsParsed.success
    ? paginationMaxResultsParsed.data
    : DEFAULT_PAGINATION_MAX_RESULTS;

  const tableColumns: TableColumnsType<ApiSchema> = [
    {
      dataIndex: "schemaType",
      ellipsis: { showTitle: false },
      key: "type",
      render: (_, { description, title, type }: ApiSchema) => (
        <Tooltip
          placement="topLeft"
          title={title && description ? `${title}: ${description}` : title || description}
        >
          <Typography.Text strong>{type}</Typography.Text>
        </Tooltip>
      ),
      sorter: {
        multiple: 1,
      },
      sortOrder: sorters.find(({ field }) => field === "schemaType")?.order,
      title: SCHEMA_TYPE,
    },
    {
      dataIndex: "schemaVersion",
      key: "version",
      render: (_, { version }: ApiSchema) => (
        <Typography.Text strong>{version || "-"}</Typography.Text>
      ),
      sorter: {
        multiple: 2,
      },
      sortOrder: sorters.find(({ field }) => field === "schemaVersion")?.order,
      title: "Schema version",
    },
    {
      dataIndex: "importDate",
      key: "createdAt",
      render: (_, { createdAt }: ApiSchema) => (
        <Typography.Text>{formatDate(createdAt)}</Typography.Text>
      ),
      sorter: {
        multiple: 3,
      },
      sortOrder: sorters.find(({ field }) => field === "importDate")?.order,
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
          maxResults: paginationMaxResults,
          page: paginationPage,
          query: queryParam || undefined,
          sorters: parseSorters(sortParam),
        },
        signal,
      });
      if (response.success) {
        setApiSchemas({ data: response.data.items.successful, status: "successful" });
        setPaginationTotal(response.data.meta.total);
        updateUrlParams({
          maxResults: response.data.meta.max_results,
          page: response.data.meta.page,
        });
        notifyParseErrors(response.data.items.failed);
      } else {
        if (!isAbortedError(response.error)) {
          setApiSchemas({ error: response.error, status: "failed" });
        }
      }
    },
    [env, queryParam, identifier, paginationMaxResults, paginationPage, sortParam, updateUrlParams]
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
              apiSchemas.status === "failed" ? (
                <ErrorResult error={apiSchemas.error.message} />
              ) : (
                <NoResults searchQuery={queryParam} />
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
          <Space size="middle">
            <Card.Meta title={SCHEMAS} />

            <Tag>{paginationTotal}</Tag>
          </Space>
        </Row>
      }
    />
  );
}
