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
import { generatePath, useNavigate } from "react-router-dom";

import {
  CredentialStatus,
  credentialStatusParser,
  getCredentials,
} from "src/adapters/api/credentials";
import { ReactComponent as IconCreditCardRefresh } from "src/assets/icons/credit-card-refresh.svg";
import { ReactComponent as IconDots } from "src/assets/icons/dots-vertical.svg";
import { ReactComponent as IconInfoCircle } from "src/assets/icons/info-circle.svg";
import { ReactComponent as IconTrash } from "src/assets/icons/trash-01.svg";
import { ReactComponent as IconClose } from "src/assets/icons/x.svg";
import { IssueDirectlyButton } from "src/components/connections/IssueDirectlyButton";
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
  DELETE,
  DETAILS,
  DID_SEARCH_PARAM,
  DOTS_DROPDOWN_WIDTH,
  EXPIRATION,
  ISSUED_CREDENTIALS,
  ISSUE_DATE,
  REVOCATION,
  REVOKE,
} from "src/utils/constants";
import { notifyParseError, notifyParseErrors } from "src/utils/error";
import { formatDate } from "src/utils/forms";

export function CredentialsTable({ userID }: { userID: string }) {
  const env = useEnvContext();

  const [credentials, setCredentials] = useState<AsyncTask<Credential[], AppError>>({
    status: "pending",
  });
  const [credentialStatus, setCredentialStatus] = useState<CredentialStatus>("all");
  const [credentialToDelete, setCredentialToDelete] = useState<Credential>();
  const [credentialToRevoke, setCredentialToRevoke] = useState<Credential>();
  const [query, setQuery] = useState<string | null>(null);

  const navigate = useNavigate();
  const navigateToDirectIssue = () =>
    navigate({
      pathname: generatePath(ROUTES.issueCredential.path),
      search: `${DID_SEARCH_PARAM}=${userID}`,
    });

  const tableColumns: ColumnsType<Credential> = [
    {
      dataIndex: "schemaType",
      ellipsis: { showTitle: false },
      key: "schemaType",
      render: (schemaType: Credential["schemaType"]) => (
        <Tooltip placement="topLeft" title={schemaType}>
          <Typography.Text strong>{schemaType}</Typography.Text>
        </Tooltip>
      ),
      sorter: ({ schemaType: a }, { schemaType: b }) => a.localeCompare(b),
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
      dataIndex: "revoked",
      key: "revoked",
      render: (revoked: Credential["revoked"]) => (
        <Typography.Text>{revoked ? "Revoked" : "-"}</Typography.Text>
      ),
      responsive: ["md"],
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

  const fetchCredentials = useCallback(
    async (signal?: AbortSignal) => {
      if (userID) {
        setCredentials((previousCredentials) =>
          isAsyncTaskDataAvailable(previousCredentials)
            ? { data: previousCredentials.data, status: "reloading" }
            : { status: "loading" }
        );
        const response = await getCredentials({
          env,
          params: {
            did: userID,
            query: query || undefined,
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
      }
    },
    [userID, env, query, credentialStatus]
  );

  const handleStatusChange = ({ target: { value } }: RadioChangeEvent) => {
    const parsedCredentialStatus = credentialStatusParser.safeParse(value);
    if (parsedCredentialStatus.success) {
      setCredentialStatus(parsedCredentialStatus.data);
    } else {
      notifyParseError(parsedCredentialStatus.error);
    }
  };

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchCredentials);

    return aborter;
  }, [fetchCredentials]);

  const credentialsList = isAsyncTaskDataAvailable(credentials) ? credentials.data : [];

  const showDefaultContent =
    credentials.status === "successful" && credentialsList.length === 0 && query === null;

  return (
    <Space direction="vertical">
      <TableCard
        defaultContents={
          <>
            <Avatar className="avatar-color-cyan" icon={<IconCreditCardRefresh />} size={48} />

            <Typography.Text strong>
              No {credentialStatus !== "all" && credentialStatus} credentials issued
            </Typography.Text>

            <Typography.Text type="secondary">
              Credentials for this connection will be listed here.
            </Typography.Text>
          </>
        }
        extraButton={<IssueDirectlyButton onClick={navigateToDirectIssue} />}
        isLoading={isAsyncTaskStarting(credentials)}
        onSearch={setQuery}
        query={query}
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
            dataSource={credentialsList}
            locale={{
              emptyText:
                credentials.status === "failed" ? (
                  <ErrorResult error={credentials.error.message} />
                ) : (
                  <NoResults searchQuery={query} />
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
              <Card.Meta title={ISSUED_CREDENTIALS} />

              <Tag color="blue">{credentialsList.length}</Tag>
            </Space>
            {showDefaultContent && credentialStatus === "all" ? (
              <IssueDirectlyButton onClick={navigateToDirectIssue} />
            ) : (
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
    </Space>
  );
}
