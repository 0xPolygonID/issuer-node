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
import { generatePath, useNavigate, useSearchParams } from "react-router-dom";

import { VerifyCredentialUser } from "../shared/VerifyCredentialUser";
import { credentialStatusParser, getCredentials } from "src/adapters/api/credentials";
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
import { ROUTES } from "src/routes";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import {
  //VERIFY_IDENTITY,
  DELETE,
  DETAILS,
  DOTS_DROPDOWN_WIDTH,
  EXPIRATION,
  EXPIRED,
  ISSUED,
  ISSUE_CREDENTIAL,
  ISSUE_DATE,
  QUERY_SEARCH_PARAM,
  REQUEST_FOR_VC,
  REVOCATION,
  REVOKE,
  REVOKE_DATE,
  STATUS_SEARCH_PARAM,
  VERIFY_IDENTITY,
} from "src/utils/constants";
import { notifyParseError, notifyParseErrors } from "src/utils/error";
import { formatDate } from "src/utils/forms";

export function CredentialsTable() {
  const env = useEnvContext();

  const navigate = useNavigate();

  const User = localStorage.getItem("user");
  const UserDID = localStorage.getItem("userId");

  const [credentials, setCredentials] = useState<AsyncTask<Credential[], AppError>>({
    status: "pending",
  });
  const [credentialToDelete, setCredentialToDelete] = useState<Credential>();
  const [credentialToRevoke, setCredentialToRevoke] = useState<Credential>();
  const [verifyCredentialForRequest, setVerifyCredentialForRequest] = useState<Credential>();

  const [searchParams, setSearchParams] = useSearchParams();

  const statusParam = searchParams.get(STATUS_SEARCH_PARAM);
  const queryParam = searchParams.get(QUERY_SEARCH_PARAM);

  const parsedStatusParam = credentialStatusParser.safeParse(statusParam);
  const credentialStatus = parsedStatusParam.success ? parsedStatusParam.data : "all";

  const credentialsList = isAsyncTaskDataAvailable(credentials) ? credentials.data : [];
  const showDefaultContent =
    credentials.status === "successful" && credentialsList.length === 0 && queryParam === null;

  let tableColumns: ColumnsType<Credential>;
  if (User === "issuer" || User === "verifier") {
    tableColumns = [
      {
        dataIndex: "userID",
        key: "userDID",
        render: (userDID: Credential["userDID"]) => (
          <Tooltip placement="topLeft" title={userDID}>
            <Typography.Text strong>{userDID}</Typography.Text>
          </Tooltip>
        ),
        title: "UserDID",
        width: "20%",
      },
      {
        dataIndex: "id",
        key: "schemaType",
        render: (schemaType: Credential["schemaType"]) => (
          <Tooltip placement="topLeft" title={schemaType}>
            <Typography.Text strong>{schemaType}</Typography.Text>
          </Tooltip>
        ),
        title: "Credential",
        width: "20%",
      },
      {
        dataIndex: "createdAt",
        key: "createdAt",
        render: (createdAt: Credential["createdAt"]) => (
          <Typography.Text>{formatDate(createdAt)}</Typography.Text>
        ),
        sorter: ({ createdAt: a }, { createdAt: b }) => dayjs(a).unix() - dayjs(b).unix(),
        title: ISSUE_DATE,
      },
      {
        dataIndex: "expired",
        key: "expired",
        render: (expired: Credential["expired"]) => (
          <Typography.Text>{expired ? "Yes" : "No"}</Typography.Text>
        ),
        responsive: ["md"],
        title: EXPIRED,
      },
      {
        dataIndex: "revoked",
        key: "revoked",
        render: (revoked: Credential["revoked"]) => (
          <Typography.Text>{revoked ? "Revoked" : "-"}</Typography.Text>
        ),
        responsive: ["md"],
        title: REVOCATION,
      },
      {
        dataIndex: "expiresAt",
        key: "expiresAt",
        render: (expiresAt: Credential["expiresAt"], credential: Credential) =>
          expiresAt ? (
            <Tooltip placement="topLeft" title={formatDate(expiresAt)}>
              <Typography.Text>
                {credential.expired ? "Expired" : dayjs(expiresAt).fromNow(true)}
              </Typography.Text>
            </Tooltip>
          ) : (
            "-"
          ),
        responsive: ["sm"],
        sorter: ({ expiresAt: a }, { expiresAt: b }) => {
          if (a && b) {
            return dayjs(a).unix() - dayjs(b).unix();
          } else if (a) {
            return -1;
          } else {
            return 1;
          }
        },
        title: EXPIRATION,
      },
      {
        dataIndex: "revokeDate",
        key: "revokeDate",
        render: (revokeDate: Credential["revokeDate"]) => (
          <Typography.Text>{revokeDate ? formatDate(revokeDate) : "-"}</Typography.Text>
        ),
        sorter: ({ revokeDate: a }, { revokeDate: b }) => dayjs(a).unix() - dayjs(b).unix(),
        title: REVOKE_DATE,
      },
      {
        dataIndex: "id",
        key: "id",
        render: (id: Credential["id"], credential: Credential) => (
          <Dropdown
            menu={{
              items:
                User === "verifier"
                  ? [
                      {
                        icon: <IconInfoCircle />,
                        key: "details",
                        label: REQUEST_FOR_VC,
                        onClick: () => setVerifyCredentialForRequest(credential),
                      },
                    ]
                  : [
                      {
                        icon: <IconInfoCircle />,
                        key: "details",
                        label: DETAILS,
                        onClick: () =>
                          navigate(
                            generatePath(ROUTES.credentialDetails.path, { credentialID: id })
                          ),
                      },
                      {
                        key: "divider1",
                        type: "divider",
                      },
                      // {
                      //   icon: <IconInfoCircle />,
                      //   key: "details",
                      //   label: VERIFY_IDENTITY,
                      // },
                      // {
                      //   icon: <IconInfoCircle />,
                      //   key: "details",
                      //   label: ISSUE_CREDENTIAL,
                      // },
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
        dataIndex: "id",
        key: "schemaType",
        render: (schemaType: Credential["schemaType"]) => (
          <Tooltip placement="topLeft" title={schemaType}>
            <Typography.Text strong>{schemaType}</Typography.Text>
          </Tooltip>
        ),
        title: "Credential",
      },
      {
        dataIndex: "createdAt",
        key: "createdAt",
        render: (createdAt: Credential["createdAt"]) => (
          <Typography.Text>{formatDate(createdAt)}</Typography.Text>
        ),
        sorter: ({ createdAt: a }, { createdAt: b }) => dayjs(a).unix() - dayjs(b).unix(),
        title: ISSUE_DATE,
      },
      {
        dataIndex: "expired",
        key: "expired",
        render: (expired: Credential["expired"]) => (
          <Typography.Text>{expired ? "Yes" : "No"}</Typography.Text>
        ),
        responsive: ["md"],
        title: EXPIRED,
      },
      {
        dataIndex: "revoked",
        key: "revoked",
        render: (revoked: Credential["revoked"]) => (
          <Typography.Text>{revoked ? "Revoked" : "-"}</Typography.Text>
        ),
        responsive: ["md"],
        title: REVOCATION,
      },
      {
        dataIndex: "expiresAt",
        key: "expiresAt",
        render: (expiresAt: Credential["expiresAt"], credential: Credential) =>
          expiresAt ? (
            <Tooltip placement="topLeft" title={formatDate(expiresAt)}>
              <Typography.Text>
                {credential.expired ? "Expired" : dayjs(expiresAt).fromNow(true)}
              </Typography.Text>
            </Tooltip>
          ) : (
            "-"
          ),
        responsive: ["sm"],
        sorter: ({ expiresAt: a }, { expiresAt: b }) => {
          if (a && b) {
            return dayjs(a).unix() - dayjs(b).unix();
          } else if (a) {
            return -1;
          } else {
            return 1;
          }
        },
        title: EXPIRATION,
      },
      {
        dataIndex: "revokeDate",
        key: "revokeDate",
        render: (revokeDate: Credential["revokeDate"]) => (
          <Typography.Text>{revokeDate ? formatDate(revokeDate) : "-"}</Typography.Text>
        ),
        sorter: ({ revokeDate: a }, { revokeDate: b }) => dayjs(a).unix() - dayjs(b).unix(),
        title: REVOKE_DATE,
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
          query: User === "issuer" || User === "verifier" ? queryParam || undefined : UserDID,
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
    [env, queryParam, credentialStatus, User, UserDID]
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

            <Typography.Text strong>No credentials</Typography.Text>

            <Typography.Text type="secondary">
              Issued credentials will be listed here.
            </Typography.Text>

            {/* {credentialStatus === "all" && (
              <Link to={generatePath(ROUTES.issueCredential.path)}>
                <Button icon={<IconCreditCardPlus />} type="primary">
                  {ISSUE_CREDENTIAL}
                </Button>
              </Link>
            )} */}
          </>
        }
        isLoading={isAsyncTaskStarting(credentials)}
        onSearch={onSearch}
        query={queryParam}
        searchPlaceholder="Search credentials, attributes, identifiers..."
        showDefaultContents={showDefaultContent}
        table={
          User !== "verifier" || queryParam ? (
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
          ) : (
            <Table />
          )
        }
        title={
          <Row gutter={[0, 8]} justify="space-between">
            <Space size="middle">
              <Card.Meta title={ISSUED} />

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
      {verifyCredentialForRequest && (
        <VerifyCredentialUser
          credential={verifyCredentialForRequest}
          onClose={() => setVerifyCredentialForRequest(undefined)}
        />
      )}
    </>
  );
}
