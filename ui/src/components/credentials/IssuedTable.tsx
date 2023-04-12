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
  message,
} from "antd";
import Table, { ColumnsType } from "antd/es/table";
import dayjs from "dayjs";
import { useCallback, useEffect, useState } from "react";
import { Link, generatePath, useSearchParams } from "react-router-dom";

import { APIError } from "src/adapters/api";
import {
  credentialStatusParser,
  deleteCredential,
  getCredentials,
  revokeCredential,
} from "src/adapters/api/credentials";
import { ReactComponent as IconCreditCardPlus } from "src/assets/icons/credit-card-plus.svg";
import { ReactComponent as IconCreditCardRefresh } from "src/assets/icons/credit-card-refresh.svg";
import { ReactComponent as IconDots } from "src/assets/icons/dots-vertical.svg";
import { ReactComponent as IconInfoCircle } from "src/assets/icons/info-circle.svg";
import { ReactComponent as IconTrash } from "src/assets/icons/trash-01.svg";
import { ReactComponent as IconClose } from "src/assets/icons/x.svg";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { NoResults } from "src/components/shared/NoResults";
import { TableCard } from "src/components/shared/TableCard";
import { useEnvContext } from "src/contexts/env";
import { Credential } from "src/domain";
import { ROUTES } from "src/routes";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import {
  CREDENTIALS,
  DOTS_DROPDOWN_WIDTH,
  EXPIRATION,
  ISSUE_CREDENTIAL,
  ISSUE_DATE,
  QUERY_SEARCH_PARAM,
  REVOCATION,
  STATUS_SEARCH_PARAM,
} from "src/utils/constants";
import { processZodError } from "src/utils/error";
import { formatDate } from "src/utils/forms";

export function IssuedTable() {
  const env = useEnvContext();

  const [credentials, setCredentials] = useState<AsyncTask<Credential[], APIError>>({
    status: "pending",
  });

  const [searchParams, setSearchParams] = useSearchParams();

  const statusParam = searchParams.get(STATUS_SEARCH_PARAM);
  const queryParam = searchParams.get(QUERY_SEARCH_PARAM);
  const parsedStatusParam = credentialStatusParser.safeParse(statusParam);
  const credentialStatus = parsedStatusParam.success ? parsedStatusParam.data : "all";

  const credentialsList = isAsyncTaskDataAvailable(credentials) ? credentials.data : [];
  const showDefaultContent =
    credentials.status === "successful" && credentialsList.length === 0 && queryParam === null;

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
      sorter: ({ attributes: { type: a } }, { attributes: { type: b } }) => a.localeCompare(b),
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
      sorter: ({ revoked: a }, { revoked: b }) => (a === b ? 0 : a ? 1 : -1),
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
                label: "Details",
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
                label: "Revoke",
                onClick: () =>
                  void revokeCredential({ env, nonce: credential.revNonce }).then((response) => {
                    if (response.isSuccessful) {
                      void fetchCredentials();

                      void message.success(response.data);
                    } else {
                      void message.error(response.error.message);
                    }
                  }),
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
                onClick: () =>
                  void deleteCredential({ env, id }).then((response) => {
                    if (response.isSuccessful) {
                      void fetchCredentials();

                      void message.success(response.data);
                    } else {
                      void message.error(response.error.message);
                    }
                  }),
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
      setCredentials((oldCredentials) =>
        isAsyncTaskDataAvailable(oldCredentials)
          ? { data: oldCredentials.data, status: "reloading" }
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
    },
    [env, queryParam, credentialStatus]
  );

  const onSearch = useCallback(
    (query: string) => {
      setSearchParams((oldParams) => {
        const oldQuery = oldParams.get(QUERY_SEARCH_PARAM);
        const params = new URLSearchParams(oldParams);

        if (query === "") {
          params.delete(QUERY_SEARCH_PARAM);
          return params;
        } else if (oldQuery !== query) {
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
      processZodError(parsedCredentialStatus.error).forEach((error) => void message.error(error));
    }
  };

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchCredentials);

    return aborter;
  }, [fetchCredentials]);

  return (
    <TableCard
      defaultContents={
        <>
          <Avatar className="avatar-color-cyan" icon={<IconCreditCardRefresh />} size={48} />

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
        <Row align="middle" justify="space-between">
          <Space align="end" size="middle">
            <Card.Meta title={ISSUE_CREDENTIAL} />

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
  );
}
