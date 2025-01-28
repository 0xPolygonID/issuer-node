import {
  App,
  Avatar,
  Button,
  Card,
  Dropdown,
  Row,
  Space,
  Table,
  TableColumnsType,
  Tag,
  Typography,
} from "antd";
import { useCallback, useEffect, useState } from "react";
import { Link, generatePath, useNavigate, useSearchParams } from "react-router-dom";

import { Sorter, parseSorters, serializeSorters } from "src/adapters/api";
import { deleteDisplayMethod, getDisplayMethods } from "src/adapters/api/display-method";
import { notifyErrors, positiveIntegerFromStringParser } from "src/adapters/parsers";
import { tableSorterParser } from "src/adapters/parsers/view";
import IconCheckMark from "src/assets/icons/check.svg?react";
import IconCopy from "src/assets/icons/copy-01.svg?react";
import IconDisplayMethod from "src/assets/icons/display-method.svg?react";
import IconDots from "src/assets/icons/dots-vertical.svg?react";
import IconInfoCircle from "src/assets/icons/info-circle.svg?react";
import IconPlus from "src/assets/icons/plus.svg?react";
import { DeleteItem } from "src/components/shared/DeleteItem";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { NoResults } from "src/components/shared/NoResults";
import { TableCard } from "src/components/shared/TableCard";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { AppError, DisplayMethod } from "src/domain";
import { ROUTES } from "src/routes";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import {
  DEFAULT_PAGINATION_MAX_RESULTS,
  DEFAULT_PAGINATION_PAGE,
  DEFAULT_PAGINATION_TOTAL,
  DETAILS,
  DISPLAY_METHOD_ADD_NEW,
  DOTS_DROPDOWN_WIDTH,
  PAGINATION_MAX_RESULTS_PARAM,
  PAGINATION_PAGE_PARAM,
  QUERY_SEARCH_PARAM,
  SORT_PARAM,
} from "src/utils/constants";

export function DisplayMethodsTable() {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();

  const { message } = App.useApp();
  const navigate = useNavigate();

  const [displayMethods, setDisplayMethods] = useState<AsyncTask<DisplayMethod[], AppError>>({
    status: "pending",
  });

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

  const displayMethodsList = isAsyncTaskDataAvailable(displayMethods) ? displayMethods.data : [];
  const showDefaultContent =
    displayMethods.status === "successful" &&
    displayMethodsList.length === 0 &&
    queryParam === null;

  const tableColumns: TableColumnsType<DisplayMethod> = [
    {
      dataIndex: "name",
      key: "name",
      render: (name: DisplayMethod["name"], { id }: DisplayMethod) => (
        <Typography.Link
          onClick={() =>
            navigate(
              generatePath(ROUTES.displayMethodDetails.path, {
                displayMethodID: id,
              })
            )
          }
          strong
        >
          {name}
        </Typography.Link>
      ),
      sorter: {
        multiple: 1,
      },
      sortOrder: sorters.find(({ field }) => field === "name")?.order,
      title: "Name",
    },
    {
      dataIndex: "type",
      key: "type",
      render: (type: DisplayMethod["type"]) => type,
      title: "Type",
    },
    {
      dataIndex: "url",
      key: "url",
      render: (url: DisplayMethod["url"]) => {
        return (
          <Typography.Link
            copyable={{
              icon: [<IconCopy key={0} />, <IconCheckMark key={1} />],
            }}
            href={url}
            target="_blank"
          >
            {url}
          </Typography.Link>
        );
      },
      title: "Url",
    },
    {
      dataIndex: "id",
      key: "id",
      render: (id: DisplayMethod["id"]) => (
        <Dropdown
          menu={{
            items: [
              {
                icon: <IconInfoCircle />,
                key: "details",
                label: DETAILS,
                onClick: () =>
                  navigate(
                    generatePath(ROUTES.displayMethodDetails.path, {
                      displayMethodID: id,
                    })
                  ),
              },
              {
                key: "divider1",
                type: "divider",
              },
              {
                danger: true,
                key: "delete",
                label: (
                  <DeleteItem
                    onOk={() => handleDeleteDisplayMethod(id)}
                    title="Are you sure you want to delete this display method?"
                  />
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

  const handleDeleteDisplayMethod = (id: string) => {
    void deleteDisplayMethod({ env, id, identifier }).then((response) => {
      if (response.success) {
        void fetchDisplayMethods();
        void message.success(response.data.message);
      } else {
        void message.error(response.error.message);
      }
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

  const fetchDisplayMethods = useCallback(
    async (signal?: AbortSignal) => {
      setDisplayMethods((previousDisplayMethods) =>
        isAsyncTaskDataAvailable(previousDisplayMethods)
          ? { data: previousDisplayMethods.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await getDisplayMethods({
        env,
        identifier,
        params: {
          maxResults: paginationMaxResults,
          page: paginationPage,
          sorters: parseSorters(sortParam),
        },
        signal,
      });
      if (response.success) {
        setDisplayMethods({
          data: response.data.items.successful,
          status: "successful",
        });
        setPaginationTotal(response.data.meta.total);
        updateUrlParams({
          maxResults: response.data.meta.max_results,
          page: response.data.meta.page,
        });
        void notifyErrors(response.data.items.failed);
      } else {
        if (!isAbortedError(response.error)) {
          setDisplayMethods({ error: response.error, status: "failed" });
        }
      }
    },
    [env, paginationMaxResults, paginationPage, sortParam, identifier, updateUrlParams]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchDisplayMethods);

    return aborter;
  }, [fetchDisplayMethods]);

  return (
    <TableCard
      defaultContents={
        <>
          <Avatar className="avatar-color-icon" icon={<IconDisplayMethod />} size={48} />

          <Typography.Text strong>No display methods</Typography.Text>

          <Typography.Text type="secondary">
            Your display methods will be listed here.
          </Typography.Text>

          <Link to={ROUTES.createDisplayMethod.path}>
            <Button icon={<IconPlus />} type="primary">
              {DISPLAY_METHOD_ADD_NEW}
            </Button>
          </Link>
        </>
      }
      isLoading={isAsyncTaskStarting(displayMethods)}
      query={queryParam}
      showDefaultContents={showDefaultContent}
      table={
        <Table
          columns={tableColumns}
          dataSource={displayMethodsList}
          locale={{
            emptyText:
              displayMethods.status === "failed" ? (
                <ErrorResult error={displayMethods.error.message} />
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
            <Card.Meta title="Display methods" />
            <Tag>{paginationTotal}</Tag>
          </Space>
        </Row>
      }
    />
  );
}
