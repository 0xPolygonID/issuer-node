import {
  Avatar,
  Button,
  Card,
  Dropdown,
  Radio,
  RadioChangeEvent,
  Row,
  Space,
  Table,
  TableColumnsType,
  Tag,
  Tooltip,
  Typography,
} from "antd";
import dayjs from "dayjs";
import { useCallback, useEffect, useState } from "react";
import { Link, generatePath, useNavigate, useSearchParams } from "react-router-dom";

import { Sorter, parseSorters, serializeSorters } from "src/adapters/api";
import { credentialStatusParser, getCredentials } from "src/adapters/api/credentials";
import {
  notifyErrors,
  notifyParseError,
  positiveIntegerFromStringParser,
} from "src/adapters/parsers";
import { tableSorterParser } from "src/adapters/parsers/view";
import IconCreditCardPlus from "src/assets/icons/credit-card-plus.svg?react";
import IconCreditCardRefresh from "src/assets/icons/credit-card-refresh.svg?react";
import IconDots from "src/assets/icons/dots-vertical.svg?react";
import IconInfoCircle from "src/assets/icons/info-circle.svg?react";
import IconTrash from "src/assets/icons/trash-01.svg?react";
import IconClose from "src/assets/icons/x.svg?react";
import { CredentialDeleteModal } from "src/components/shared/CredentialDeleteModal";
import { CredentialRevokeModal } from "src/components/shared/CredentialRevokeModal";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { NoResults } from "src/components/shared/NoResults";
import { TableCard } from "src/components/shared/TableCard";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { AppError, Credential } from "src/domain";
import { ROUTES } from "src/routes";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import {
  DEFAULT_PAGINATION_MAX_RESULTS,
  DEFAULT_PAGINATION_PAGE,
  DEFAULT_PAGINATION_TOTAL,
  DELETE,
  DETAILS,
  DOTS_DROPDOWN_WIDTH,
  EXPIRATION,
  ISSUED,
  ISSUE_CREDENTIAL,
  ISSUE_DATE,
  PAGINATION_MAX_RESULTS_PARAM,
  PAGINATION_PAGE_PARAM,
  QUERY_SEARCH_PARAM,
  REVOCATION,
  REVOKE,
  SORT_PARAM,
  STATUS_SEARCH_PARAM,
} from "src/utils/constants";
import { formatDate } from "src/utils/forms";

export function CredentialsTable() {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();

  const navigate = useNavigate();

  const [credentials, setCredentials] = useState<AsyncTask<Credential[], AppError>>({
    status: "pending",
  });
  const [credentialToDelete, setCredentialToDelete] = useState<Credential>();
  const [credentialToRevoke, setCredentialToRevoke] = useState<Credential>();

  const [searchParams, setSearchParams] = useSearchParams();

  const statusParam = searchParams.get(STATUS_SEARCH_PARAM);
  const queryParam = searchParams.get(QUERY_SEARCH_PARAM);
  const paginationPageParam = searchParams.get(PAGINATION_PAGE_PARAM);
  const paginationMaxResultsParam = searchParams.get(PAGINATION_MAX_RESULTS_PARAM);
  const sortParam = searchParams.get(SORT_PARAM);

  const sorters = parseSorters(sortParam);
  const parsedStatusParam = credentialStatusParser.safeParse(statusParam);
  const credentialStatus = parsedStatusParam.success ? parsedStatusParam.data : "all";
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

  const credentialsList = isAsyncTaskDataAvailable(credentials) ? credentials.data : [];

  const showDefaultContent =
    credentials.status === "successful" && credentialsList.length === 0 && queryParam === null;

  const tableColumns: TableColumnsType<Credential> = [
    {
      dataIndex: "schemaType",
      ellipsis: { showTitle: false },
      key: "schemaType",
      render: (schemaType: Credential["schemaType"], credential: Credential) => (
        <Typography.Link
          onClick={() =>
            navigate(generatePath(ROUTES.credentialDetails.path, { credentialID: credential.id }))
          }
          strong
        >
          {schemaType}
        </Typography.Link>
      ),
      sorter: {
        multiple: 1,
      },
      sortOrder: sorters.find(({ field }) => field === "schemaType")?.order,
      title: "Credential",
    },
    {
      dataIndex: "issuanceDate",
      key: "createdAt",
      render: (_, { issuanceDate }: Credential) => (
        <Typography.Text>{formatDate(issuanceDate)}</Typography.Text>
      ),
      sorter: {
        multiple: 2,
      },
      sortOrder: sorters.find(({ field }) => field === "createdAt")?.order,
      title: ISSUE_DATE,
    },
    {
      dataIndex: "expirationDate",
      key: "expiresAt",
      render: (expirationDate: Credential["expirationDate"], credential: Credential) =>
        expirationDate ? (
          <Tooltip placement="topLeft" title={formatDate(expirationDate)}>
            <Typography.Text>
              {credential.expired ? "Expired" : dayjs(expirationDate).fromNow(true)}
            </Typography.Text>
          </Tooltip>
        ) : (
          "-"
        ),
      responsive: ["md"],
      sorter: {
        multiple: 3,
      },
      sortOrder: sorters.find(({ field }) => field === "expiresAt")?.order,
      title: EXPIRATION,
    },
    {
      dataIndex: "revoked",
      key: "revoked",
      render: (revoked: Credential["revoked"]) => (
        <Typography.Text>{revoked ? "Revoked" : "-"}</Typography.Text>
      ),
      responsive: ["sm"],
      sorter: {
        multiple: 4,
      },
      sortOrder: sorters.find(({ field }) => field === "revoked")?.order,
      title: REVOCATION,
    },
    {
      dataIndex: "id",
      key: "id",
      render: (id: Credential["id"], credential: Credential) => (
        <Dropdown
          menu={{
            items: [
              {
                icon: <IconInfoCircle />,
                key: "details",
                label: DETAILS,
                onClick: () =>
                  navigate(
                    generatePath(ROUTES.credentialDetails.path, {
                      credentialID: id,
                    })
                  ),
              },
              {
                key: "divider1",
                type: "divider",
              },
              {
                danger: true,
                disabled: credential.revoked,
                icon: <IconClose />,
                key: "revoke",
                label: REVOKE,
                onClick: () => setCredentialToRevoke(credential),
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
                onClick: () => setCredentialToDelete(credential),
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
        const fieldMap: Record<string, string> = {
          issuanceDate: "createdAt",
          expirationDate: "expiresAt",
        };
        const newSorters = (sorters || parseSorters(sortParam)).map((s) => ({
          ...s,
          field: fieldMap[s.field] ?? s.field,
        }));
        newSorters.length > 0
          ? params.set(SORT_PARAM, serializeSorters(newSorters))
          : params.delete(SORT_PARAM);

        return params;
      });
    },
    [setSearchParams, sortParam]
  );

  const fetchCredentials = useCallback(
    async (signal?: AbortSignal) => {
      setCredentials((previousCredentials) =>
        isAsyncTaskDataAvailable(previousCredentials)
          ? { data: previousCredentials.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await getCredentials({
        env,
        identifier,
        params: {
          maxResults: paginationMaxResults,
          page: paginationPage,
          query: queryParam || undefined,
          sorters: parseSorters(sortParam),
          status: credentialStatus,
        },
        signal,
      });
      if (response.success) {
        setCredentials({
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
          setCredentials({ error: response.error, status: "failed" });
        }
      }
    },
    [
      credentialStatus,
      env,
      paginationMaxResults,
      paginationPage,
      queryParam,
      sortParam,
      identifier,
      updateUrlParams,
    ]
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

  const handleStatusChange = ({ target: { value } }: RadioChangeEvent) => {
    const parsedCredentialStatus = credentialStatusParser.safeParse(value);
    if (parsedCredentialStatus.success) {
      const params = new URLSearchParams(searchParams);

      if (parsedCredentialStatus.data === "all") {
        params.delete(STATUS_SEARCH_PARAM);
      } else {
        params.set(STATUS_SEARCH_PARAM, parsedCredentialStatus.data);
      }

      setSearchParams(params);
    } else {
      void notifyParseError(parsedCredentialStatus.error);
    }
  };

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchCredentials);

    return aborter;
  }, [fetchCredentials]);

  return (
    <>
      <TableCard
        defaultContents={
          <>
            <Avatar className="avatar-color-icon" icon={<IconCreditCardRefresh />} size={48} />

            <Typography.Text strong>No credentials</Typography.Text>

            <Typography.Text type="secondary">
              Issued credentials will be listed here.
            </Typography.Text>

            {credentialStatus === "all" && (
              <Link to={generatePath(ROUTES.issueCredential.path)}>
                <Button icon={<IconCreditCardPlus />} type="primary">
                  {ISSUE_CREDENTIAL}
                </Button>
              </Link>
            )}
          </>
        }
        isLoading={isAsyncTaskStarting(credentials)}
        onSearch={onSearch}
        query={queryParam}
        searchPlaceholder="Search credentials, attributes, identifiers..."
        showDefaultContents={showDefaultContent}
        table={
          <Table
            columns={tableColumns}
            dataSource={credentialsList}
            loading={credentials.status === "reloading"}
            locale={{
              emptyText:
                credentials.status === "failed" ? (
                  <ErrorResult error={credentials.error.message} />
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
            tableLayout="fixed"
          />
        }
        title={
          <Row gutter={[0, 8]} justify="space-between">
            <Space size="middle">
              <Card.Meta title={ISSUED} />

              <Tag>{paginationTotal}</Tag>
            </Space>

            {(!showDefaultContent || credentialStatus !== "all") && (
              <Radio.Group onChange={handleStatusChange} value={credentialStatus}>
                <Radio.Button value="all">All</Radio.Button>

                <Radio.Button value="revoked">Revoked</Radio.Button>

                <Radio.Button value="expired">Expired</Radio.Button>
              </Radio.Group>
            )}
          </Row>
        }
      />
      {credentialToDelete && (
        <CredentialDeleteModal
          credential={credentialToDelete}
          onClose={() => setCredentialToDelete(undefined)}
          onDelete={() => void fetchCredentials()}
        />
      )}
      {credentialToRevoke && (
        <CredentialRevokeModal
          credential={credentialToRevoke}
          onClose={() => setCredentialToRevoke(undefined)}
          onRevoke={() => void fetchCredentials()}
        />
      )}
    </>
  );
}
