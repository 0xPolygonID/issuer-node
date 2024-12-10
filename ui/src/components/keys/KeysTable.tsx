import {
  Avatar,
  Button,
  Card,
  Dropdown,
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

import { getKeys } from "src/adapters/api/keys";
import { positiveIntegerFromStringParser } from "src/adapters/parsers";
import IconIssuers from "src/assets/icons/building-08.svg?react";
import IconCheckMark from "src/assets/icons/check.svg?react";
import IconCopy from "src/assets/icons/copy-01.svg?react";
import IconDots from "src/assets/icons/dots-vertical.svg?react";
import IconInfoCircle from "src/assets/icons/info-circle.svg?react";
import IconPlus from "src/assets/icons/plus.svg?react";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { NoResults } from "src/components/shared/NoResults";
import { TableCard } from "src/components/shared/TableCard";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { AppError, Key } from "src/domain";
import { ROUTES } from "src/routes";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import {
  DEFAULT_PAGINATION_MAX_RESULTS,
  DEFAULT_PAGINATION_PAGE,
  DEFAULT_PAGINATION_TOTAL,
  DETAILS,
  DOTS_DROPDOWN_WIDTH,
  KEY_ADD_NEW,
  PAGINATION_MAX_RESULTS_PARAM,
  PAGINATION_PAGE_PARAM,
  QUERY_SEARCH_PARAM,
} from "src/utils/constants";
import { notifyParseErrors } from "src/utils/error";

export function KeysTable() {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();

  const navigate = useNavigate();

  const [keys, setKeys] = useState<AsyncTask<Key[], AppError>>({
    status: "pending",
  });

  const [searchParams, setSearchParams] = useSearchParams();
  const queryParam = searchParams.get(QUERY_SEARCH_PARAM);
  const paginationPageParam = searchParams.get(PAGINATION_PAGE_PARAM);
  const paginationMaxResultsParam = searchParams.get(PAGINATION_MAX_RESULTS_PARAM);

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

  const keysList = isAsyncTaskDataAvailable(keys) ? keys.data : [];
  const showDefaultContent = keys.status === "successful" && keysList.length === 0;

  const tableColumns: TableColumnsType<Key> = [
    {
      dataIndex: "publicKey",
      key: "publicKey",
      render: (publicKey: Key["publicKey"]) => (
        <Tooltip title={publicKey}>
          <Typography.Text
            copyable={{
              icon: [<IconCopy key={0} />, <IconCheckMark key={1} />],
            }}
            ellipsis={{
              suffix: publicKey.slice(-5),
            }}
            strong
          >
            {publicKey}
          </Typography.Text>
        </Tooltip>
      ),
      title: "Public key",
    },
    {
      dataIndex: "keyType",
      key: "keyType",
      render: (keyType: Key["keyType"]) => <Typography.Text>{keyType}</Typography.Text>,
      title: "Type",
    },
    {
      dataIndex: "isAuthCoreClaim",
      key: "isAuthCoreClaim",
      render: (isAuthCoreClaim: Key["isAuthCoreClaim"]) => (
        <Typography.Text>{`${isAuthCoreClaim}`}</Typography.Text>
      ),
      title: "Auth core claim",
    },
    {
      dataIndex: "id",
      key: "id",
      render: (id: Key["id"]) => (
        <Dropdown
          menu={{
            items: [
              {
                icon: <IconInfoCircle />,
                key: "details",
                label: DETAILS,
                onClick: () =>
                  navigate(
                    generatePath(ROUTES.keyDetails.path, {
                      keyID: id,
                    })
                  ),
              },
            ],
          }}
        >
          <Row>
            <IconDots className="icon-secondary" />
          </Row>
        </Dropdown>
      ),
      width: DOTS_DROPDOWN_WIDTH,
    },
  ];

  const updateUrlParams = useCallback(
    ({ maxResults, page }: { maxResults?: number; page?: number }) => {
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

        return params;
      });
    },
    [setSearchParams]
  );

  const fetchKeys = useCallback(
    async (signal?: AbortSignal) => {
      setKeys((previousKeys) =>
        isAsyncTaskDataAvailable(previousKeys)
          ? { data: previousKeys.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await getKeys({
        env,
        identifier,
        params: {
          maxResults: paginationMaxResults,
          page: paginationPage,
        },
        signal,
      });
      if (response.success) {
        setKeys({
          data: response.data.items.successful,
          status: "successful",
        });
        setPaginationTotal(response.data.meta.total);
        updateUrlParams({
          maxResults: response.data.meta.max_results,
          page: response.data.meta.page,
        });
        notifyParseErrors(response.data.items.failed);
      } else {
        if (!isAbortedError(response.error)) {
          setKeys({ error: response.error, status: "failed" });
        }
      }
    },
    [env, paginationMaxResults, paginationPage, identifier, updateUrlParams]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchKeys);

    return aborter;
  }, [fetchKeys]);

  return (
    <TableCard
      defaultContents={
        <>
          <Avatar className="avatar-color-icon" icon={<IconIssuers />} size={48} />

          <Typography.Text strong>No keys</Typography.Text>

          <Typography.Text type="secondary">Your keys will be listed here.</Typography.Text>

          <Link to={ROUTES.createKey.path}>
            <Button icon={<IconPlus />} type="primary">
              {KEY_ADD_NEW}
            </Button>
          </Link>
        </>
      }
      isLoading={isAsyncTaskStarting(keys)}
      query={queryParam}
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
          dataSource={keysList}
          locale={{
            emptyText:
              keys.status === "failed" ? (
                <ErrorResult error={keys.error.message} />
              ) : (
                <NoResults searchQuery={queryParam} />
              ),
          }}
          onChange={({ current, pageSize, total }) => {
            setPaginationTotal(total || DEFAULT_PAGINATION_TOTAL);
            updateUrlParams({
              maxResults: pageSize,
              page: current,
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
            <Card.Meta title="Display methods" />
            <Tag>{paginationTotal}</Tag>
          </Space>
        </Row>
      }
    />
  );
}
