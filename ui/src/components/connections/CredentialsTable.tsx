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
import EncryptedIcon from "src/assets/icons/key-01.svg?react";
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
      dataIndex: "type",
      key: "type",
      render: (type: Credential["type"]) =>
        type === "encrypted" ? (
          <Tooltip placement="topLeft" title="Encrypted credential">
            <Avatar
              className="avatar-color-icon"
              icon={<EncryptedIcon />}
              size={24}
              style={{ backgroundColor: "transparent", padding: 1 }}
            />
          </Tooltip>
        ) : null,
      responsive: ["sm"],
      title: "",
      width: 48,
    },
    {
      dataIndex: ["data", "schemaType"],
      ellipsis: { showTitle: false },
      key: "schemaType",
      render: (schemaType: Credential["data"]["schemaType"], credential: Credential) => (
        <Typography.Link
          onClick={() =>
            navigate(
              generatePath(ROUTES.credentialDetails.path, { credentialID: credential.data.id })
            )
          }
          strong
        >
          {schemaType}
        </Typography.Link>
      ),

      sorter: ({ data: { schemaType: a } }, { data: { schemaType: b } }) => a.localeCompare(b),
      title: "Credential",
    },
    {
      dataIndex: ["data", "issuanceDate"],
      key: "issuanceDate",
      render: (issuanceDate: Credential["data"]["issuanceDate"]) => (
        <Typography.Text>{formatDate(issuanceDate)}</Typography.Text>
      ),
      sorter: ({ data: { issuanceDate: a } }, { data: { issuanceDate: b } }) =>
        dayjs(a).unix() - dayjs(b).unix(),
      title: ISSUE_DATE,
    },
    {
      dataIndex: ["data", "expirationDate"],
      key: "expirationDate",
      render: (expirationDate: Credential["data"]["expirationDate"], credential: Credential) =>
        expirationDate ? (
          <Tooltip placement="topLeft" title={formatDate(expirationDate)}>
            <Typography.Text>
              {credential.data.expired ? "Expired" : dayjs(expirationDate).fromNow(true)}
            </Typography.Text>
          </Tooltip>
        ) : (
          "-"
        ),
      responsive: ["sm"],
      sorter: ({ data: { expirationDate: a } }, { data: { expirationDate: b } }) => {
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
      dataIndex: ["data", "revoked"],
      key: "revoked",
      render: (revoked: Credential["data"]["revoked"]) => (
        <Typography.Text>{revoked ? "Revoked" : "-"}</Typography.Text>
      ),
      responsive: ["md"],
      title: REVOCATION,
    },
    {
      dataIndex: ["data", "id"],
      key: "id",
      render: (id: Credential["data"]["id"], credential: Credential) => (
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
                disabled: credential.data.revoked,
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
            columns={tableColumns}
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
            rowKey={(credential) => credential.data.id}
            showSorterTooltip
            sortDirections={["ascend", "descend"]}
            tableLayout="fixed"
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
