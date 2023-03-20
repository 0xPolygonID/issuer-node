import {
  Avatar,
  Card,
  Radio,
  RadioChangeEvent,
  Row,
  Space,
  Switch,
  Table,
  Tag,
  Tooltip,
  Typography,
  message,
} from "antd";
import { ColumnsType } from "antd/es/table";
import { useCallback, useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { z } from "zod";

import { Credential, credentialUpdate, credentialsGetAll } from "src/adapters/api/credentials";
import { Schema } from "src/adapters/api/schemas";
import { ReactComponent as IconDots } from "src/assets/icons/dots-vertical.svg";
import { ReactComponent as IconLink } from "src/assets/icons/link-03.svg";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { NoResults } from "src/components/shared/NoResults";
import { TableCard } from "src/components/shared/TableCard";
import { useEnvContext } from "src/contexts/env";
import { APIError, processZodError } from "src/utils/adapters";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { ACCESSIBLE_UNTIL, LINKS, QUERY_SEARCH_PARAM } from "src/utils/constants";
import { formatDate } from "src/utils/forms";
import {
  AsyncTask,
  StrictSchema,
  isAsyncTaskDataAvailable,
  isAsyncTaskStarting,
} from "src/utils/types";

const SHOW_SEARCH_PARAM = "show";

type Show = "all" | "active" | "inactive" | "exceeded";

const showParser = StrictSchema<Show>()(
  z.union([z.literal("all"), z.literal("active"), z.literal("inactive"), z.literal("exceeded")])
);

export function LinkTable() {
  const env = useEnvContext();
  const [credentials, setCredentials] = useState<AsyncTask<Credential[], APIError>>({
    status: "pending",
  });
  const [isCredentialUpdating, setCredentialUpdating] = useState<Record<string, boolean>>({});

  const [searchParams, setSearchParams] = useSearchParams();

  const credentialsList = isAsyncTaskDataAvailable(credentials) ? credentials.data : [];
  const showParam = searchParams.get(SHOW_SEARCH_PARAM);
  const queryParam = searchParams.get(QUERY_SEARCH_PARAM);
  const parsedShowParam = showParser.safeParse(showParam);

  const show = parsedShowParam.success ? parsedShowParam.data : "all";

  const tableContents: ColumnsType<Credential> = [
    {
      dataIndex: "active",
      ellipsis: true,
      key: "active",
      render: (active: boolean, credential: Credential) => (
        <Switch
          checked={credential.valid && active}
          disabled={!credential.valid}
          loading={isCredentialUpdating[credential.id]}
          onClick={(isActive, event) => {
            event.stopPropagation();
            toggleCredentialActive(isActive, credential);
          }}
          size="small"
        />
      ),
      sorter: ({ active: a }, { active: b }) => (a === b ? 0 : a ? 1 : -1),
      title: "Active",
      width: 100,
    },
    {
      dataIndex: "schemaTemplate",
      ellipsis: true,
      key: "schemaTemplate",
      render: ({ schema }: Schema) => (
        <Tooltip placement="topLeft" title={schema}>
          <Typography.Text strong>{schema}</Typography.Text>
        </Tooltip>
      ),
      sorter: ({ schemaTemplate: { schema: a } }, { schemaTemplate: { schema: b } }) =>
        a.localeCompare(b),
      title: "Credential",
    },
    {
      dataIndex: "linkAccessibleUntil",
      ellipsis: true,
      key: "linkAccessibleUntil",
      render: (date: Date | null) => (
        <Typography.Text>{date ? formatDate(date, true) : "Unlimited"}</Typography.Text>
      ),
      title: ACCESSIBLE_UNTIL,
    },
    {
      dataIndex: "linkCurrentIssuance",
      ellipsis: true,
      key: "linkCurrentIssuance",
      render: (issued: number | null) => {
        const value = issued ? issued : 0;

        return <Typography.Text>{value}</Typography.Text>;
      },
      title: "Credentials issued",
    },
    {
      dataIndex: "linkMaximumIssuance",
      ellipsis: true,
      key: "linkMaximumIssuance",
      render: (linkMaximumIssuance: number | null) => {
        const value = linkMaximumIssuance ? linkMaximumIssuance : "Unlimited";

        return <Typography.Text>{value}</Typography.Text>;
      },
      title: "Maximum issuance",
    },
    {
      dataIndex: "active",
      key: "active",
      render: (active: boolean, credential: Credential) => (
        <Row justify="space-between">
          {credential.valid ? (
            active ? (
              <Tag color="success">Active</Tag>
            ) : (
              <Tag>Inactive</Tag>
            )
          ) : (
            <Tag color="error">Exceeded</Tag>
          )}
          <IconDots className="icon-secondary" />
        </Row>
      ),
      title: "Status",
      width: 140,
    },
  ];

  const getCredentials = useCallback(
    async (signal: AbortSignal) => {
      setCredentials((previousCredentials) =>
        isAsyncTaskDataAvailable(previousCredentials)
          ? { data: previousCredentials.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await credentialsGetAll({
        env,
        params: {
          query: queryParam || undefined,
          valid: show === "exceeded" ? false : undefined,
        },
        signal,
      });

      if (response.isSuccessful) {
        setCredentials({ data: response.data.credentials, status: "successful" });

        response.data.errors.forEach((zodError) => {
          processZodError(zodError).forEach((error) => void message.error(error));
        });
      } else {
        if (!isAbortedError(response.error)) {
          setCredentials({ error: response.error, status: "failed" });
        }
      }
    },
    [env, queryParam, show]
  );

  const handleShowChange = ({ target: { value } }: RadioChangeEvent) => {
    const parsedShowValue = showParser.safeParse(value);
    if (parsedShowValue.success) {
      const params = new URLSearchParams(searchParams);
      if (parsedShowValue.data === "exceeded") {
        params.set(SHOW_SEARCH_PARAM, parsedShowValue.data);
      } else {
        params.delete(SHOW_SEARCH_PARAM);
      }

      setCredentials({ status: "pending" });
      setSearchParams(params);
    } else {
      processZodError(parsedShowValue.error).forEach((error) => void message.error(error));
    }
  };

  const updateCredentialInState = (credential: Credential) => {
    setCredentials((oldCredentials) =>
      isAsyncTaskDataAvailable(oldCredentials)
        ? {
            data: oldCredentials.data.map((currentCredential: Credential) =>
              currentCredential.id === credential.id ? credential : currentCredential
            ),
            status: "successful",
          }
        : oldCredentials
    );
  };

  const toggleCredentialActive = (
    active: boolean,
    { id: credentialID, schemaTemplate }: Credential
  ) => {
    setCredentialUpdating((currentCredentialsUpdating) => {
      return { ...currentCredentialsUpdating, [credentialID]: true };
    });

    void credentialUpdate({
      credentialID,
      env,
      payload: { active },
      schemaID: schemaTemplate.id,
    }).then((response) => {
      if (response.isSuccessful) {
        updateCredentialInState(response.data);

        void message.success(`Link ${active ? "activated" : "deactivated"}`);
      } else {
        void message.error(response.error.message);
      }

      setCredentialUpdating((currentCredentialsUpdating) => {
        return { ...currentCredentialsUpdating, [credentialID]: false };
      });
    });
  };

  useEffect(() => {
    const { aborter } = makeRequestAbortable(getCredentials);

    return aborter;
  }, [getCredentials]);

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

  return (
    <>
      <TableCard
        defaultContents={
          <>
            <Avatar className="avatar-color-cyan" icon={<IconLink />} size={48} />

            <Typography.Text strong>No links</Typography.Text>

            <Typography.Text type="secondary">
              Credential links will be listed here.
            </Typography.Text>
          </>
        }
        isLoading={isAsyncTaskStarting(credentials)}
        onSearch={onSearch}
        query={queryParam}
        searchPlaceholder="Search credentials, attributes..."
        showDefaultContents={
          credentials.status === "successful" && credentialsList.length === 0 && queryParam === null
        }
        table={
          <Table
            columns={tableContents.map(({ title, ...column }) => ({
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
          <Row justify="space-between">
            <Space size="middle">
              <Card.Meta title={LINKS} />

              <Tag color="blue">{credentialsList.length}</Tag>
            </Space>

            <Radio.Group onChange={handleShowChange} value={show}>
              <Radio.Button value="all">All</Radio.Button>

              <Radio.Button value="exceeded">Exceeded</Radio.Button>
            </Radio.Group>
          </Row>
        }
      />
    </>
  );
}
