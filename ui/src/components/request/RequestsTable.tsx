import {
  Avatar,
  Button,
  Card,
  Dropdown,
  Radio,
  RadioChangeEvent,
  Row,
  Space,
  Tag,
  Tooltip,
  Typography,
} from "antd";
import Table, { ColumnsType } from "antd/es/table";
import dayjs from "dayjs";
import { useCallback, useEffect, useState } from "react";
import { Link, generatePath, useNavigate, useSearchParams } from "react-router-dom";

import { credentialStatusParser, getCredentials } from "src/adapters/api/credentials";
import { ReactComponent as IconCreditCardPlus } from "src/assets/icons/credit-card-plus.svg";
import { ReactComponent as IconCreditCardRefresh } from "src/assets/icons/credit-card-refresh.svg";
import { ReactComponent as IconDots } from "src/assets/icons/dots-vertical.svg";
import { ReactComponent as IconInfoCircle } from "src/assets/icons/info-circle.svg";
import { ReactComponent as IconTrash } from "src/assets/icons/trash-01.svg";
import { ReactComponent as IconClose } from "src/assets/icons/x.svg";
import { CredentialDeleteModal } from "src/components/shared/CredentialDeleteModal";
import { CredentialRevokeModal } from "src/components/shared/CredentialRevokeModal";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { NoResults } from "src/components/shared/NoResults";
import { TableCard } from "src/components/shared/TableCard";
import { useEnvContext } from "src/contexts/Env";
import { AppError, Credential } from "src/domain";
import { Request } from "src/domain/request";
import { ROUTES } from "src/routes";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import {
  APPROVE1,
  APPROVE2,
  DELETE,
  DETAILS,
  DOTS_DROPDOWN_WIDTH,
  ISSUE_REQUEST,
  QUERY_SEARCH_PARAM,
  REQUEST_DATE,
  REVOKE,
  STATUS_SEARCH_PARAM,
} from "src/utils/constants";
import { notifyParseError, notifyParseErrors } from "src/utils/error";
import { formatDate } from "src/utils/forms";

export function RequestsTable() {
  const env = useEnvContext();

  const navigate = useNavigate();

  const User = localStorage.getItem("user");

  const [credentials, setCredentials] = useState<AsyncTask<Credential[], AppError>>({
    status: "pending",
  });
  const [credentialToDelete, setCredentialToDelete] = useState<Credential>();
  const [credentialToRevoke, setCredentialToRevoke] = useState<Credential>();

  const [searchParams, setSearchParams] = useSearchParams();

  const statusParam = searchParams.get(STATUS_SEARCH_PARAM);
  const queryParam = searchParams.get(QUERY_SEARCH_PARAM);
  const parsedStatusParam = credentialStatusParser.safeParse(statusParam);
  const credentialStatus = parsedStatusParam.success ? parsedStatusParam.data : "all";

  const credentialsList = isAsyncTaskDataAvailable(credentials) ? credentials.data : [];
  const showDefaultContent =
    credentials.status === "successful" && credentialsList.length === 0 && queryParam === null;

  let tableColumns: ColumnsType<Request>;
  if (User === "verifier") {
    tableColumns = [
      {
        dataIndex: "requestId",
        key: "requestId",
        render: (requestId: Request["requestId"]) => (
          <Tooltip placement="topLeft" title={requestId}>
            <Typography.Text strong>{requestId}</Typography.Text>
          </Tooltip>
        ),
        title: "Request Id",
      },
      {
        dataIndex: "requestType",
        key: "requestType",
        render: (requestType: Request["requestType"]) => (
          <Tooltip placement="topLeft" title={requestType}>
            <Typography.Text strong>{requestType}</Typography.Text>
          </Tooltip>
        ),
        title: "Request Type",
      },
      {
        dataIndex: "status",
        key: "status",
        render: (status: Request["status"]) => (
          <Tooltip placement="topLeft" title={status}>
            <Typography.Text strong>{status}</Typography.Text>
          </Tooltip>
        ),
        title: "Status",
      },
      {
        dataIndex: "requestDate",
        key: "requestDate",
        render: (requestDate: Request["requestDate"]) => (
          <Typography.Text>{formatDate(requestDate)}</Typography.Text>
        ),
        sorter: ({ requestDate: a }, { requestDate: b }) => dayjs(a).unix() - dayjs(b).unix(),
        title: REQUEST_DATE,
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
                    navigate(generatePath(ROUTES.credentialDetails.path, { credentialID: id })),
                },
                {
                  key: "divider1",
                  type: "divider",
                },
                {
                  icon: <IconInfoCircle />,
                  key: "details",
                  label: APPROVE1,
                },
                {
                  icon: <IconInfoCircle />,
                  key: "details",
                  label: APPROVE2,
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
  } else if (User === "issuer") {
    tableColumns = [
      {
        dataIndex: "requestId",
        key: "requestId",
        render: (requestId: Request["requestId"]) => (
          <Tooltip placement="topLeft" title={requestId}>
            <Typography.Text strong>{requestId}</Typography.Text>
          </Tooltip>
        ),
        title: "Request Id",
      },
      {
        dataIndex: "userDID",
        key: "userDID",
        render: (userDID: Request["userDID"]) => (
          <Tooltip placement="topLeft" title={userDID}>
            <Typography.Text strong>{userDID}</Typography.Text>
          </Tooltip>
        ),
        title: "UserDID",
      },
      {
        dataIndex: "credentialType",
        key: "credentialType",
        render: (credentialType: Request["credentialType"]) => (
          <Tooltip placement="topLeft" title={credentialType}>
            <Typography.Text strong>{credentialType}</Typography.Text>
          </Tooltip>
        ),
        title: "Credential Type",
      },
      {
        dataIndex: "requestType",
        key: "requestType",
        render: (requestType: Request["requestType"]) => (
          <Tooltip placement="topLeft" title={requestType}>
            <Typography.Text strong>{requestType}</Typography.Text>
          </Tooltip>
        ),
        title: "Request Type",
      },
      {
        dataIndex: "status",
        key: "status",
        render: (status: Request["status"]) => (
          <Tooltip placement="topLeft" title={status}>
            <Typography.Text strong>{status}</Typography.Text>
          </Tooltip>
        ),
        title: "Status",
      },
      {
        dataIndex: "requestDate",
        key: "requestDate",
        render: (requestDate: Request["requestDate"]) => (
          <Typography.Text>{formatDate(requestDate)}</Typography.Text>
        ),
        sorter: ({ requestDate: a }, { requestDate: b }) => dayjs(a).unix() - dayjs(b).unix(),
        title: REQUEST_DATE,
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
                    navigate(generatePath(ROUTES.credentialDetails.path, { credentialID: id })),
                },
                {
                  key: "divider1",
                  type: "divider",
                },
                {
                  icon: <IconInfoCircle />,
                  key: "details",
                  label: APPROVE1,
                },
                {
                  icon: <IconInfoCircle />,
                  key: "details",
                  label: APPROVE2,
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
  } else {
    tableColumns = [
      {
        dataIndex: "requestId",
        key: "requestId",
        render: (requestId: Request["requestId"]) => (
          <Tooltip placement="topLeft" title={requestId}>
            <Typography.Text strong>{requestId}</Typography.Text>
          </Tooltip>
        ),
        title: "Request Id",
      },
      {
        dataIndex: "userDID",
        key: "userDID",
        render: (userDID: Request["userDID"]) => (
          <Tooltip placement="topLeft" title={userDID}>
            <Typography.Text strong>{userDID}</Typography.Text>
          </Tooltip>
        ),
        title: "UserDID",
      },
      {
        dataIndex: "status",
        key: "status",
        render: (status: Request["status"]) => (
          <Tooltip placement="topLeft" title={status}>
            <Typography.Text strong>{status}</Typography.Text>
          </Tooltip>
        ),
        title: "Status",
      },
      {
        dataIndex: "requestDate",
        key: "requestDate",
        render: (requestDate: Request["requestDate"]) => (
          <Typography.Text>{formatDate(requestDate)}</Typography.Text>
        ),
        sorter: ({ requestDate: a }, { requestDate: b }) => dayjs(a).unix() - dayjs(b).unix(),
        title: REQUEST_DATE,
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
                    navigate(generatePath(ROUTES.credentialDetails.path, { credentialID: id })),
                },
                {
                  key: "divider1",
                  type: "divider",
                },
                {
                  icon: <IconInfoCircle />,
                  key: "details",
                  label: APPROVE1,
                },
                {
                  icon: <IconInfoCircle />,
                  key: "details",
                  label: APPROVE2,
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
  }

  const fetchCredentials = useCallback(
    async (signal?: AbortSignal) => {
      setCredentials((previousCredentials) =>
        isAsyncTaskDataAvailable(previousCredentials)
          ? { data: previousCredentials.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await getCredentials({
        env,
        params: {
          query: queryParam || undefined,
          status: credentialStatus,
        },
        signal,
      });
      if (response.success) {
        setCredentials({
          data: response.data.successful,
          status: "successful",
        });
        notifyParseErrors(response.data.failed);
      } else {
        if (!isAbortedError(response.error)) {
          setCredentials({ error: response.error, status: "failed" });
        }
      }
    },
    [env, queryParam, credentialStatus]
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
      notifyParseError(parsedCredentialStatus.error);
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
            <Avatar className="avatar-color-cyan" icon={<IconCreditCardRefresh />} size={48} />

            <Typography.Text strong>No Requests</Typography.Text>

            <Typography.Text type="secondary">Issued Request will be listed here.</Typography.Text>

            {credentialStatus === "all" && (
              <Link to={generatePath(ROUTES.issueCredential.path)}>
                <Button icon={<IconCreditCardPlus />} type="primary">
                  {ISSUE_REQUEST}
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
            columns={tableColumns.map(({ title, ...column }) => ({
              title: (
                <Typography.Text type="secondary">
                  <>{title}</>
                </Typography.Text>
              ),
              ...column,
            }))}
            dataSource={credentialsList}
            locale={{
              emptyText:
                credentials.status === "failed" ? (
                  <ErrorResult error={credentials.error.message} />
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
          <Row gutter={[0, 8]} justify="space-between">
            <Space size="middle">
              <Card.Meta title={ISSUE_REQUEST} />

              <Tag color="blue">{credentialsList.length}</Tag>
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
