import {
  Avatar,
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
import { useSearchParams } from "react-router-dom";

import { IssueCredentialUser } from "../shared/IssueCredentialUser";
import { RequestDeleteModal } from "../shared/RequestDeleteModal";
import { RequestRevokeModal } from "../shared/RequestRevokeModal";
import { getRequests, requestStatusParser } from "src/adapters/api/requests";
import { ReactComponent as IconCreditCardRefresh } from "src/assets/icons/credit-card-refresh.svg";
import { ReactComponent as IconDots } from "src/assets/icons/dots-vertical.svg";
import { ReactComponent as IconInfoCircle } from "src/assets/icons/info-circle.svg";
import { ReactComponent as IconTrash } from "src/assets/icons/trash-01.svg";
import { ReactComponent as IconClose } from "src/assets/icons/x.svg";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { NoResults } from "src/components/shared/NoResults";
import { TableCard } from "src/components/shared/TableCard";
import { useEnvContext } from "src/contexts/Env";
import { AppError } from "src/domain";
import { Request } from "src/domain/request";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import {
  DELETE,
  DOTS_DROPDOWN_WIDTH,
  ISSUE_CREDENTIAL,
  ISSUE_REQUEST,
  QUERY_SEARCH_PARAM,
  REQUEST_DATE,
  REVOKE,
  STATUS_SEARCH_PARAM,
  VERIFY_IDENTITY,
} from "src/utils/constants";
import { notifyParseError, notifyParseErrors } from "src/utils/error";
import { formatDate } from "src/utils/forms";

export function RequestsTable() {
  const env = useEnvContext();

  const User = localStorage.getItem("user");

  const [requests, setRequests] = useState<AsyncTask<Request[], AppError>>({
    status: "pending",
  });
  const [requestToDelete, setRequestToDelete] = useState<Request>();
  const [requestToRevoke, setRequestToRevoke] = useState<Request>();
  const [issueCredentialForRequest, setIssueCredentialForRequest] = useState<Request>();

  const [searchParams, setSearchParams] = useSearchParams();

  const statusParam = searchParams.get(STATUS_SEARCH_PARAM);
  const queryParam = searchParams.get(QUERY_SEARCH_PARAM);

  const parsedStatusParam = requestStatusParser.safeParse(statusParam);
  const requestStatus = parsedStatusParam.success ? parsedStatusParam.data : "all";

  const requestsList = isAsyncTaskDataAvailable(requests) ? requests.data : [];
  const showDefaultContent =
    requests.status === "successful" && requestsList.length === 0 && queryParam === null;

  let tableColumns: ColumnsType<Request>;
  if (User === "verifier" || User === "issuer") {
    tableColumns = [
      {
        dataIndex: "id",
        key: "requestId",
        render: (requestId: Request["requestId"]) => (
          <Tooltip placement="topLeft" title={requestId}>
            <Typography.Text strong>{requestId}</Typography.Text>
          </Tooltip>
        ),
        title: "Request ID",
        width: "20%",
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
        width: "20%",
      },
      {
        dataIndex: "credential_type",
        key: "credentialType",
        render: (credentialType: Request["credentialType"]) => (
          <Tooltip placement="topLeft" title={credentialType}>
            <Typography.Text strong>{credentialType}</Typography.Text>
          </Tooltip>
        ),
        title: "Credential Type",
      },
      {
        dataIndex: "request_type",
        key: "requestType",
        render: (requestType: Request["requestType"]) => (
          <Tooltip placement="topLeft" title={requestType}>
            <Typography.Text strong>{requestType}</Typography.Text>
          </Tooltip>
        ),
        title: "Request Type",
      },
      {
        dataIndex: "Active",
        key: "status",
        render: (status: Request["status"]) => (
          <Tooltip placement="topLeft" title={status}>
            <Typography.Text strong>{status ? "Active" : "-"}</Typography.Text>
          </Tooltip>
        ),
        title: "Status",
      },
      {
        dataIndex: "created_at",
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
        render: (id: Request["id"], request: Request) => (
          <Dropdown
            menu={{
              items: [
                // {
                //   icon: <IconInfoCircle />,
                //   key: "details",
                //   label: DETAILS,
                //   onClick: () =>
                //     navigate(generatePath(ROUTES.credentialDetails.path, { credentialID: id })),
                // },
                {
                  key: "divider1",
                  type: "divider",
                },
                {
                  disabled: request.Active,
                  icon: <IconInfoCircle />,
                  key: "verify",
                  label: VERIFY_IDENTITY,
                },
                {
                  icon: <IconInfoCircle />,
                  key: "issue",
                  label: ISSUE_CREDENTIAL,
                  onClick: () => setIssueCredentialForRequest(request),
                },
                {
                  danger: true,
                  disabled: request.Active,
                  icon: <IconClose />,
                  key: "revoke",
                  label: REVOKE,
                  onClick: () => setRequestToRevoke(request),
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
                  onClick: () => setRequestToDelete(request),
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
        dataIndex: "Active",
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
        render: (id: Request["id"], request: Request) => (
          <Dropdown
            menu={{
              items: [
                // {
                //   icon: <IconInfoCircle />,
                //   key: "details",
                //   label: DETAILS,
                //   onClick: () =>
                //     navigate(generatePath(ROUTES.credentialDetails.path, { credentialID: id })),
                // },
                {
                  key: "divider1",
                  type: "divider",
                },
                {
                  disabled: request.Active,
                  icon: <IconInfoCircle />,
                  key: "verify",
                  label: VERIFY_IDENTITY,
                },
                {
                  icon: <IconInfoCircle />,
                  key: "issue",
                  label: ISSUE_CREDENTIAL,
                },
                {
                  danger: true,
                  disabled: request.Active,
                  icon: <IconClose />,
                  key: "revoke",
                  label: REVOKE,
                  onClick: () => setRequestToRevoke(request),
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
                  onClick: () => setRequestToDelete(request),
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

  const fetchRequests = useCallback(
    async (signal?: AbortSignal) => {
      setRequests((previousRequests) =>
        isAsyncTaskDataAvailable(previousRequests)
          ? { data: previousRequests.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await getRequests({
        env,
        params: {
          query: queryParam || undefined,
          status: requestStatus,
        },
        signal,
      });
      if (response.success) {
        setRequests({
          data: response.data.successful,
          status: "successful",
        });
        notifyParseErrors(response.data.failed);
      } else {
        if (!isAbortedError(response.error)) {
          setRequests({ error: response.error, status: "failed" });
        }
      }
    },
    [env, queryParam, requestStatus]
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
    const parsedRequestStatus = requestStatusParser.safeParse(value);
    if (parsedRequestStatus.success) {
      const params = new URLSearchParams(searchParams);

      if (parsedRequestStatus.data === "all") {
        params.delete(STATUS_SEARCH_PARAM);
      } else {
        params.set(STATUS_SEARCH_PARAM, parsedRequestStatus.data);
      }

      setSearchParams(params);
    } else {
      notifyParseError(parsedRequestStatus.error);
    }
  };

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchRequests);

    return aborter;
  }, [fetchRequests]);

  return (
    <>
      <TableCard
        defaultContents={
          <>
            <Avatar className="avatar-color-cyan" icon={<IconCreditCardRefresh />} size={48} />

            <Typography.Text strong>No Requests</Typography.Text>

            <Typography.Text type="secondary">Issued Request will be listed here.</Typography.Text>

            {/* {requestStatus === "all" && (
              <Link to={generatePath(ROUTES.issueCredential.path)}>
                <Button icon={<IconCreditCardPlus />} type="primary">
                  {ISSUE_REQUEST}
                </Button>
              </Link>
            )} */}
          </>
        }
        isLoading={isAsyncTaskStarting(requests)}
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
            dataSource={requestsList}
            locale={{
              emptyText:
                requests.status === "failed" ? (
                  <ErrorResult error={requests.error.message} />
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

              <Tag color="blue">{requestsList.length}</Tag>
            </Space>

            {(!showDefaultContent || requestStatus !== "all") && (
              <Radio.Group onChange={handleStatusChange} value={requestStatus}>
                <Radio.Button value="all">All</Radio.Button>

                <Radio.Button value="revoked">Revoked</Radio.Button>

                <Radio.Button value="expired">Expired</Radio.Button>
              </Radio.Group>
            )}
          </Row>
        }
      />
      {requestToDelete && (
        <RequestDeleteModal
          onClose={() => setRequestToDelete(undefined)}
          onDelete={() => void fetchRequests()}
          request={requestToDelete}
        />
      )}
      {requestToRevoke && (
        <RequestRevokeModal
          onClose={() => setRequestToRevoke(undefined)}
          onRevoke={() => void fetchRequests()}
          request={requestToRevoke}
        />
      )}
      {issueCredentialForRequest && (
        <IssueCredentialUser
          onClose={() => setIssueCredentialForRequest(undefined)}
          request={issueCredentialForRequest}
        />
      )}
    </>
  );
}
