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
  message,
} from "antd";
import Table, { ColumnsType } from "antd/es/table";
import dayjs, { extend as extendDayJsWith } from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import { useCallback, useEffect, useState } from "react";

import { APIError } from "src/adapters/api";
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
import { ErrorResult } from "src/components/shared/ErrorResult";
import { NoResults } from "src/components/shared/NoResults";
import { TableCard } from "src/components/shared/TableCard";
import { useEnvContext } from "src/contexts/env";
import { Credential } from "src/domain";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import {
  CREDENTIALS,
  DOTS_DROPDOWN_WIDTH,
  ISSUE_CREDENTIAL,
  ISSUE_DATE,
} from "src/utils/constants";
import { processZodError } from "src/utils/error";
import { formatDate } from "src/utils/forms";

extendDayJsWith(relativeTime);

export function CredentialsTable({ userID }: { userID: string }) {
  const env = useEnvContext();

  const [credentials, setCredentials] = useState<AsyncTask<Credential[], APIError>>({
    status: "pending",
  });
  const [credentialStatus, setCredentialStatus] = useState<CredentialStatus>("all");
  const [query, setQuery] = useState<string | null>(null);

  const tableColumns: ColumnsType<Credential> = [
    {
      dataIndex: "attributes",
      ellipsis: { showTitle: false },
      key: "attributes",
      render: (attributes: Credential["attributes"]) => (
        <Tooltip placement="topLeft" title={attributes.type}>
          <Typography.Text strong>{attributes.type}</Typography.Text>
        </Tooltip>
      ),
      sorter: ({ id: a }, { id: b }) => a.localeCompare(b),
      title: CREDENTIALS,
    },
    {
      dataIndex: "createdAt",
      key: "createdAt",
      render: (createdAt: Credential["createdAt"]) => (
        <Typography.Text>{formatDate(createdAt, true)}</Typography.Text>
      ),
      sorter: ({ createdAt: a }, { createdAt: b }) => a.getTime() - b.getTime(),
      title: ISSUE_DATE,
    },
    {
      dataIndex: "expiresAt",
      key: "expiresAt",
      render: (expiresAt: Credential["expiresAt"], credential: Credential) =>
        expiresAt ? (
          <Tooltip placement="topLeft" title={formatDate(expiresAt, true)}>
            <Typography.Text>
              {credential.expired ? "Expired" : dayjs(expiresAt).fromNow(true)}
            </Typography.Text>
          </Tooltip>
        ) : (
          "-"
        ),
      sorter: ({ expiresAt: a }, { expiresAt: b }) => {
        if (a && b) {
          return a.getTime() - b.getTime();
        } else if (a) {
          return 1;
        } else {
          return -1;
        }
      },
      title: "Expiration",
    },
    {
      dataIndex: "revoked",
      key: "revoked",
      render: (revoked: Credential["revoked"]) => (
        <Typography.Text>{revoked ? "Revoked" : "-"}</Typography.Text>
      ),
      title: "Revocation",
    },
    {
      dataIndex: "id",
      key: "id",
      render: () => (
        <Dropdown
          menu={{
            items: [
              {
                icon: <IconInfoCircle />,
                key: "details",
                label: "Details",
              },
              {
                key: "divider1",
                type: "divider",
              },
              {
                danger: true,
                icon: <IconClose />,
                key: "revoke",
                label: "Revoke",
              },
              {
                key: "divider2",
                type: "divider",
              },
              {
                danger: true,
                icon: <IconTrash />,
                key: "delete",
                label: "Delete",
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
    async (signal: AbortSignal) => {
      if (userID) {
        setCredentials({ status: "loading" });
        const response = await getCredentials({
          env,
          params: {
            // TODO should change when PID-498 is done
            query: query ? `${userID} ${query}` : `${userID}`,
            status: credentialStatus,
          },
          signal,
        });
        if (response.isSuccessful) {
          setCredentials({
            data: response.data,
            status: "successful",
          });
        } else {
          if (!isAbortedError(response.error)) {
            setCredentials({ error: response.error, status: "failed" });
          }
        }
      }
    },
    [userID, env, query, credentialStatus]
  );

  const handleTypeChange = ({ target: { value } }: RadioChangeEvent) => {
    const parsedCredentialType = credentialStatusParser.safeParse(value);
    if (parsedCredentialType.success) {
      setCredentialStatus(parsedCredentialType.data);
    } else {
      processZodError(parsedCredentialType.error).forEach((error) => void message.error(error));
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
      extraButton={<IssueDirectlyButton />}
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
        <Row align="middle" justify="space-between">
          <Space align="end" size="middle">
            <Card.Meta title={ISSUE_CREDENTIAL} />

            <Tag color="blue">{credentialsList.length}</Tag>
          </Space>
          {showDefaultContent && credentialStatus === "all" ? (
            <IssueDirectlyButton />
          ) : (
            <Radio.Group onChange={handleTypeChange} value={credentialStatus}>
              <Radio.Button value="all">All</Radio.Button>

              <Radio.Button value="revoked">Revoked</Radio.Button>

              <Radio.Button value="expired">Expired</Radio.Button>
            </Radio.Group>
          )}
        </Row>
      }
    />
  );
}
