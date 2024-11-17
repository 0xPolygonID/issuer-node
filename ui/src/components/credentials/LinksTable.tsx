import {
  App,
  Avatar,
  Button,
  Card,
  Dropdown,
  Grid,
  Radio,
  RadioChangeEvent,
  Row,
  Space,
  Switch,
  Table,
  TableColumnsType,
  Tag,
  Tooltip,
  Typography,
} from "antd";

import { useCallback, useEffect, useState } from "react";
import { generatePath, useNavigate, useSearchParams } from "react-router-dom";
import { Sorter, parseSorters, serializeSorters } from "src/adapters/api";

import { getLinks, linkStatusParser, updateLink } from "src/adapters/api/credentials";
import { positiveIntegerFromStringParser } from "src/adapters/parsers";
import { tableSorterParser } from "src/adapters/parsers/view";
import IconCreditCardPlus from "src/assets/icons/credit-card-plus.svg?react";
import IconDots from "src/assets/icons/dots-vertical.svg?react";
import IconInfoCircle from "src/assets/icons/info-circle.svg?react";
import IconLink from "src/assets/icons/link-03.svg?react";
import IconTrash from "src/assets/icons/trash-01.svg?react";
import { LinkDeleteModal } from "src/components/credentials/LinkDeleteModal";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { NoResults } from "src/components/shared/NoResults";
import { TableCard } from "src/components/shared/TableCard";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { AppError, Link } from "src/domain";
import { ROUTES } from "src/routes";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import {
  ACCESSIBLE_UNTIL,
  DEFAULT_PAGINATION_MAX_RESULTS,
  DEFAULT_PAGINATION_PAGE,
  DEFAULT_PAGINATION_TOTAL,
  DELETE,
  DETAILS,
  LINKS,
  PAGINATION_MAX_RESULTS_PARAM,
  PAGINATION_PAGE_PARAM,
  QUERY_SEARCH_PARAM,
  SORT_PARAM,
  STATUS,
  STATUS_SEARCH_PARAM,
} from "src/utils/constants";
import { notifyParseErrors } from "src/utils/error";
import { formatDate } from "src/utils/forms";

export function LinksTable() {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();

  const { md, sm } = Grid.useBreakpoint();
  const { message } = App.useApp();
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  const [links, setLinks] = useState<AsyncTask<Link[], AppError>>({
    status: "pending",
  });
  const [isLinkUpdating, setLinkUpdating] = useState<Record<string, boolean>>({});
  const [linkToDelete, setLinkToDelete] = useState<string>();
  const [paginationTotal, setPaginationTotal] = useState<number>(DEFAULT_PAGINATION_TOTAL);

  const linksList = isAsyncTaskDataAvailable(links) ? links.data : [];
  const statusParam = searchParams.get(STATUS_SEARCH_PARAM);
  const queryParam = searchParams.get(QUERY_SEARCH_PARAM);
  const paginationPageParam = searchParams.get(PAGINATION_PAGE_PARAM);
  const paginationMaxResultsParam = searchParams.get(PAGINATION_MAX_RESULTS_PARAM);
  const sortParam = searchParams.get(SORT_PARAM);

  const sorters = parseSorters(sortParam);

  const parsedStatusParam = linkStatusParser.safeParse(statusParam);
  const paginationPageParsed = positiveIntegerFromStringParser.safeParse(paginationPageParam);
  const paginationMaxResultsParsed =
    positiveIntegerFromStringParser.safeParse(paginationMaxResultsParam);

  const paginationPage = paginationPageParsed.success
    ? paginationPageParsed.data
    : DEFAULT_PAGINATION_PAGE;
  const paginationMaxResults = paginationMaxResultsParsed.success
    ? paginationMaxResultsParsed.data
    : DEFAULT_PAGINATION_MAX_RESULTS;

  const showDefaultContent =
    links.status === "successful" && linksList.length === 0 && queryParam === null;

  const status = parsedStatusParam.success ? parsedStatusParam.data : undefined;

  const tableColumns: TableColumnsType<Link> = [
    {
      dataIndex: "active",
      ellipsis: true,
      key: "active",
      render: (active: Link["active"], link: Link) => (
        <Switch
          checked={active && link.status !== "exceeded"}
          disabled={link.status === "exceeded"}
          loading={isLinkUpdating[link.id]}
          onClick={(isActive) => {
            toggleLinkActive(isActive, link.id);
          }}
          size="small"
        />
      ),
      sorter: {
        multiple: 1,
      },
      sortOrder: sorters.find(({ field }) => field === "active")?.order,
      title: "Active",
      width: md ? 100 : 60,
    },
    {
      dataIndex: "schemaType",
      ellipsis: true,
      key: "schemaType",
      render: (schemaType: Link["schemaType"]) => (
        <Tooltip placement="topLeft" title={schemaType}>
          <Typography.Text strong>{schemaType}</Typography.Text>
        </Tooltip>
      ),
      sorter: {
        multiple: 2,
      },
      sortOrder: sorters.find(({ field }) => field === "schemaType")?.order,
      title: "Credential",
    },
    {
      dataIndex: "accessibleUntil",
      ellipsis: true,
      key: "expiration",
      render: (_, { expiration }: Link) => (
        <Typography.Text>{expiration ? formatDate(expiration) : "Unlimited"}</Typography.Text>
      ),
      responsive: ["sm"],
      sorter: {
        multiple: 3,
      },
      sortOrder: sorters.find(({ field }) => field === "accessibleUntil")?.order,
      title: ACCESSIBLE_UNTIL,
    },
    {
      dataIndex: "credentialIssued",
      ellipsis: true,
      key: "issuedClaims",
      render: (_, { issuedClaims }: Link) => {
        const value = issuedClaims ? issuedClaims : 0;
        return <Typography.Text>{value}</Typography.Text>;
      },
      responsive: ["md"],
      sorter: {
        multiple: 4,
      },
      sortOrder: sorters.find(({ field }) => field === "credentialIssued")?.order,
      title: "Credentials issued",
    },
    {
      dataIndex: "maximumIssuance",
      ellipsis: true,
      key: "maxIssuance",
      render: (_, { maxIssuance }: Link) => {
        const value = maxIssuance ? maxIssuance : "Unlimited";

        return <Typography.Text>{value}</Typography.Text>;
      },
      responsive: ["md"],
      sorter: {
        multiple: 5,
      },
      sortOrder: sorters.find(({ field }) => field === "maximumIssuance")?.order,
      title: "Maximum issuance",
    },
    {
      dataIndex: "status",
      key: "status",
      render: (status: Link["status"], { id }: Link) => (
        <Row justify="space-between">
          {(() => {
            switch (status) {
              case "active": {
                return <Tag color="success">Active</Tag>;
              }
              case "inactive": {
                return <Tag>Inactive</Tag>;
              }
              case "exceeded": {
                return <Tag color="error">Exceeded</Tag>;
              }
            }
          })()}
          <Dropdown
            menu={{
              items: [
                {
                  icon: <IconInfoCircle />,
                  key: "details",
                  label: DETAILS,
                  onClick: () => navigate(generatePath(ROUTES.linkDetails.path, { linkID: id })),
                },
                {
                  key: "divider",
                  type: "divider",
                },
                {
                  danger: true,
                  icon: <IconTrash />,
                  key: "delete",
                  label: DELETE,
                  onClick: () => setLinkToDelete(id),
                },
              ],
            }}
          >
            <Row>
              <IconDots className="icon-secondary" />
            </Row>
          </Dropdown>
        </Row>
      ),
      sorter: {
        multiple: 6,
      },
      sortOrder: sorters.find(({ field }) => field === "status")?.order,
      title: STATUS,
      width: 140,
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

  const fetchLinks = useCallback(
    async (signal?: AbortSignal) => {
      setLinks((previousLinks) =>
        isAsyncTaskDataAvailable(previousLinks)
          ? { data: previousLinks.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await getLinks({
        env,
        identifier,
        params: {
          maxResults: paginationMaxResults,
          page: paginationPage,
          query: queryParam || undefined,
          sorters: parseSorters(sortParam),
          status: status,
        },
        signal,
      });

      if (response.success) {
        setLinks({ data: response.data.items.successful, status: "successful" });
        setPaginationTotal(response.data.meta.total);
        updateUrlParams({
          maxResults: response.data.meta.max_results,
          page: response.data.meta.page,
        });

        notifyParseErrors(response.data.items.failed);
      } else {
        if (!isAbortedError(response.error)) {
          setLinks({ error: response.error, status: "failed" });
        }
      }
    },
    [
      env,
      queryParam,
      status,
      identifier,
      paginationMaxResults,
      paginationPage,
      sortParam,
      updateUrlParams,
    ]
  );

  const handleStatusChange = ({ target: { value } }: RadioChangeEvent) => {
    const parsedLinkValue = linkStatusParser.safeParse(value);
    const params = new URLSearchParams(searchParams);

    if (parsedLinkValue.success) {
      params.set(STATUS_SEARCH_PARAM, parsedLinkValue.data);
    } else {
      params.delete(STATUS_SEARCH_PARAM);
    }
    setSearchParams(params);
    setLinks({ status: "pending" });
  };

  const updateCredentialInState = (active: Link["active"], id: Link["id"]) => {
    setLinks((previousLinks) =>
      isAsyncTaskDataAvailable(previousLinks)
        ? {
            data: previousLinks.data.reduce((links: Link[], currentLink: Link) => {
              if (currentLink.id === id) {
                if (status === currentLink.status) {
                  return links;
                } else {
                  const linkStatusInverted: Link = {
                    ...currentLink,
                    active,
                    status: currentLink.status === "active" ? "inactive" : "active",
                  };

                  return [...links, linkStatusInverted];
                }
              } else {
                return [...links, currentLink];
              }
            }, []),
            status: "successful",
          }
        : previousLinks
    );
  };

  const toggleLinkActive = (active: boolean, id: Link["id"]) => {
    setLinkUpdating((currentLinksUpdating) => {
      return { ...currentLinksUpdating, [id]: true };
    });

    void updateLink({
      env,
      id,
      identifier,
      payload: { active },
    }).then((response) => {
      if (response.success) {
        updateCredentialInState(active, id);

        void message.success(response.data.message);
      } else {
        void message.error(response.error.message);
      }

      setLinkUpdating((currentLinksUpdating) => {
        return { ...currentLinksUpdating, [id]: false };
      });
    });
  };

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchLinks);

    return aborter;
  }, [fetchLinks]);

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
        return previousParams;
      });
    },
    [setSearchParams]
  );

  return (
    <>
      <TableCard
        defaultContents={
          <>
            <Avatar className="avatar-color-icon" icon={<IconLink />} size={48} />

            <Typography.Text strong>No links</Typography.Text>

            <Typography.Text type="secondary">
              Credential links will be listed here.
            </Typography.Text>

            {status === undefined && (
              <Button
                icon={<IconCreditCardPlus />}
                onClick={() => navigate(generatePath(ROUTES.issueCredential.path))}
                type="primary"
              >
                Issue credential
              </Button>
            )}
          </>
        }
        isLoading={isAsyncTaskStarting(links)}
        onSearch={onSearch}
        query={queryParam}
        searchPlaceholder="Search credentials, attributes..."
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
            dataSource={linksList}
            locale={{
              emptyText:
                links.status === "failed" ? (
                  <ErrorResult error={links.error.message} />
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
          <Row gutter={[0, 8]} justify="space-between">
            <Space size="middle">
              <Card.Meta title={LINKS} />

              <Tag>{paginationTotal}</Tag>
            </Space>

            {(!showDefaultContent || status !== undefined) && (
              <Radio.Group onChange={handleStatusChange} value={status}>
                <Radio.Button value={undefined}>All</Radio.Button>

                <Radio.Button value="active">Active</Radio.Button>

                {/* //TODO PID-702 Merge in one button */}
                {sm && <Radio.Button value="inactive">Inactive</Radio.Button>}

                <Radio.Button value="exceeded">Exceeded</Radio.Button>
              </Radio.Group>
            )}
          </Row>
        }
      />

      {linkToDelete && (
        <LinkDeleteModal
          id={linkToDelete}
          onClose={() => setLinkToDelete(undefined)}
          onDelete={() => void fetchLinks()}
        />
      )}
    </>
  );
}
