import {
  Avatar,
  Button,
  Card,
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
import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";

import { APIError } from "src/adapters/api";
import { Connection, getConnection } from "src/adapters/api/connections";
import {
  Credential,
  CredentialType,
  getCredentials,
  typeParser,
} from "src/adapters/api/credentials";
import { ReactComponent as IconCreditCardRefresh } from "src/assets/icons/credit-card-refresh.svg";
import { ReactComponent as IconTrash } from "src/assets/icons/trash-01.svg";
import { ConnectionDetailsRowDropdown } from "src/components/connections/ConnectionDetailsRowDropdown";
import { IssueDirectlyButton } from "src/components/connections/IssueDirectlyButton";
import { Detail } from "src/components/shared/Detail";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { NoResults } from "src/components/shared/NoResults";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { TableCard } from "src/components/shared/TableCard";
import { useEnvContext } from "src/contexts/env";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import {
  CREDENTIALS,
  DOTS_DROPDOWN_WIDTH,
  IDENTIFIER,
  ISSUE_CREDENTIAL,
  ISSUE_DATE,
} from "src/utils/constants";
import { processZodError } from "src/utils/error";
import { formatDate } from "src/utils/forms";
import {
  AsyncTask,
  hasAsyncTaskFailed,
  isAsyncTaskDataAvailable,
  isAsyncTaskStarting,
} from "src/utils/types";

const obfuscateDID = (did: string) => {
  const didSplit = did.split(":").slice(0, -1).join(":");
  const address = did.split(":").pop();

  return address
    ? `${didSplit}${address.substring(0, 5)}...${address.substring(address.length - 5)}`
    : "-";
};

const timeToExpire = (date: Date): string => {
  const difference = date.getTime() - new Date().getTime();

  if (difference < 0) {
    return "Expired";
  }

  const days = Math.floor(difference / (1000 * 60 * 60 * 24));
  const hours = Math.floor((difference / (1000 * 60 * 60)) % 24);
  const minutes = Math.floor((difference / (1000 * 60)) % 60);

  if (days > 0) {
    return `${days}d ${hours}h`;
  } else if (hours > 0) {
    return `${hours}h ${minutes}m`;
  } else {
    return `${minutes}m`;
  }
};

export function ConnectionDetails() {
  const env = useEnvContext();

  const [connection, setConnection] = useState<AsyncTask<Connection, APIError>>({
    status: "pending",
  });
  const [credentials, setCredentials] = useState<AsyncTask<Credential[], APIError>>({
    status: "pending",
  });
  const [credentialType, setCredentialType] = useState<CredentialType>("all");
  const [query, setQuery] = useState<string | null>(null);

  const { connectionID } = useParams();

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
      render: (expiresAt: Credential["expiresAt"]) => (
        <Typography.Text>{expiresAt ? timeToExpire(expiresAt) : "-"}</Typography.Text>
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
      render: () => <ConnectionDetailsRowDropdown />,
      width: DOTS_DROPDOWN_WIDTH,
    },
  ];

  const fetchConnection = useCallback(
    async (signal: AbortSignal) => {
      if (connectionID) {
        setConnection({ status: "loading" });
        const response = await getConnection({
          env,
          id: connectionID,
          signal,
        });
        if (response.isSuccessful) {
          setConnection({
            data: response.data,
            status: "successful",
          });
        } else {
          if (!isAbortedError(response.error)) {
            setConnection({ error: response.error, status: "failed" });
          }
        }
      }
    },
    [connectionID, env]
  );

  const fetchCredentials = useCallback(
    async (signal: AbortSignal) => {
      if (isAsyncTaskDataAvailable(connection)) {
        setCredentials({ status: "loading" });
        const { userID } = connection.data;
        const response = await getCredentials({
          env,
          params: {
            // TODO should change when PID-498 is done
            query: query ? `${userID} ${query}` : `${userID}`,
            type: credentialType,
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
    [connection, env, query, credentialType]
  );

  const handleTypeChange = ({ target: { value } }: RadioChangeEvent) => {
    const parsedTypeValue = typeParser.safeParse(value);
    if (parsedTypeValue.success) {
      setCredentialType(parsedTypeValue.data);
    } else {
      processZodError(parsedTypeValue.error).forEach((error) => void message.error(error));
    }
  };

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchConnection);

    return aborter;
  }, [fetchConnection]);

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchCredentials);

    return aborter;
  }, [fetchCredentials]);

  const credentialsList = isAsyncTaskDataAvailable(credentials)
    ? credentials.data
    : isAsyncTaskDataAvailable(connection) && !hasAsyncTaskFailed(credentials)
    ? connection.data.credentials
    : [];

  const showDefaultContent =
    credentials.status === "successful" && credentialsList.length === 0 && query === null;

  return (
    <SiderLayoutContent
      description="View connection information, credential attribute data. Revoke and delete issued credentials."
      showBackButton
      showDivider
      title="Connection details"
    >
      <Space direction="vertical" size="large">
        <Card>
          {(() => {
            switch (connection.status) {
              case "pending":
              case "loading": {
                return <LoadingResult />;
              }
              case "failed": {
                return <ErrorResult error={connection.error.message} />;
              }
              case "successful":
              case "reloading": {
                return (
                  <Space direction="vertical" size="middle">
                    <Row align="middle" justify="space-between">
                      <Card.Meta title="Connection" />
                      <Button danger icon={<IconTrash />} type="text">
                        Delete connection
                      </Button>
                    </Row>
                    <Card className="background-grey">
                      <Detail
                        copyable
                        data={obfuscateDID(connection.data.userID)}
                        label={IDENTIFIER}
                      />
                      <Detail
                        data={formatDate(connection.data.createdAt, true)}
                        label="Creation date"
                      />
                    </Card>
                  </Space>
                );
              }
            }
          })()}
        </Card>
        {isAsyncTaskDataAvailable(connection) && (
          <TableCard
            defaultContents={
              <>
                <Avatar className="avatar-color-cyan" icon={<IconCreditCardRefresh />} size={48} />

                <Typography.Text strong>
                  No {credentialType !== "all" && credentialType} credentials issued
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
                {showDefaultContent && credentialType === "all" && <IssueDirectlyButton />}
                <Radio.Group onChange={handleTypeChange} value={credentialType}>
                  <Radio.Button value="all">All</Radio.Button>

                  <Radio.Button value="revoked">Revoked</Radio.Button>

                  <Radio.Button value="expired">Expired</Radio.Button>
                </Radio.Group>
              </Row>
            }
          />
        )}
      </Space>
    </SiderLayoutContent>
  );
}
