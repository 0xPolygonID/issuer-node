import {
  Avatar,
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
import { generatePath, useNavigate } from "react-router-dom";

import {
  CredentialStatus,
  credentialStatusParser,
  getCredentials,
} from "src/adapters/api/credentials";
import { notifyErrors, notifyParseError } from "src/adapters/parsers";
import IconCreditCardRefresh from "src/assets/icons/credit-card-refresh.svg?react";
import IconDots from "src/assets/icons/dots-vertical.svg?react";
import IconInfoCircle from "src/assets/icons/info-circle.svg?react";
import IconTrash from "src/assets/icons/trash-01.svg?react";
import IconClose from "src/assets/icons/x.svg?react";
import { IssueDirectlyButton } from "src/components/connections/IssueDirectlyButton";
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
import { formatDate } from "src/utils/forms";

export function CredentialsTable({ userID }: { userID: string }) {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();

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

      sorter: ({ schemaType: a }, { schemaType: b }) => a.localeCompare(b),
      title: "Credential",
    },
    {
      dataIndex: "issuanceDate",
      key: "issuanceDate",
      render: (issuanceDate: Credential["issuanceDate"]) => (
        <Typography.Text>{formatDate(issuanceDate)}</Typography.Text>
      ),
      sorter: ({ issuanceDate: a }, { issuanceDate: b }) => dayjs(a).unix() - dayjs(b).unix(),
      title: ISSUE_DATE,
    },
    {
      dataIndex: "expirationDate",
      key: "expirationDate",
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
      responsive: ["sm"],
      sorter: ({ expirationDate: a }, { expirationDate: b }) => {
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
          identifier,
          params: {
            credentialSubject: userID,
            query: query || undefined,
            status: credentialStatus,
          },
          signal,
        });
        if (response.success) {
          setCredentials({
            data: response.data.items.successful,
            status: "successful",
          });
          void notifyErrors(response.data.items.failed);
        } else {
          if (!isAbortedError(response.error)) {
            setCredentials({ error: response.error, status: "failed" });
          }
        }
      }
    },
    [userID, env, query, credentialStatus, identifier]
  );

  const handleStatusChange = ({ target: { value } }: RadioChangeEvent) => {
    const parsedCredentialStatus = credentialStatusParser.safeParse(value);
    if (parsedCredentialStatus.success) {
      setCredentialStatus(parsedCredentialStatus.data);
    } else {
      void notifyParseError(parsedCredentialStatus.error);
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
            <Avatar className="avatar-color-icon" icon={<IconCreditCardRefresh />} size={48} />

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

              <Tag>{credentialsList.length}</Tag>
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
