import { Avatar, Button, Card, Row, Space, Table, Tag, Tooltip, Typography, message } from "antd";
import { ColumnsType } from "antd/es/table";
import { useCallback, useEffect, useState } from "react";
import { Link, generatePath, useSearchParams } from "react-router-dom";

import { Schema, schemasGetAll, schemasUpdate } from "src/adapters/api/schemas";
import { ReactComponent as IconArchive } from "src/assets/icons/archive.svg";
import { ReactComponent as IconCertificate } from "src/assets/icons/certificate-01.svg";
import { ReactComponent as IconUpload } from "src/assets/icons/upload-01.svg";
import { ErrorResult } from "src/components/schemas/ErrorResult";
import { NoResults } from "src/components/schemas/NoResults";
import { SchemaDetails } from "src/components/schemas/SchemaDetails";
import { SchemaRowDropdown } from "src/components/schemas/SchemaRowDropdown";
import { TableCard } from "src/components/schemas/TableCard";
import { ROUTES } from "src/routes";
import { APIError, processZodError } from "src/utils/adapters";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { FORM_LABEL, QUERY_SEARCH_PARAM, SCHEMA_ID_SEARCH_PARAM } from "src/utils/constants";
import { formatDate } from "src/utils/forms";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/types";

export function MySchemas({ showActive }: { showActive: boolean }) {
  const [schemas, setSchemas] = useState<AsyncTask<Schema[], APIError>>({
    status: "pending",
  });

  const [searchParams, setSearchParams] = useSearchParams();
  const queryParam = searchParams.get(QUERY_SEARCH_PARAM);

  const getSchemas = useCallback(
    async (signal: AbortSignal) => {
      setSchemas((oldState) =>
        isAsyncTaskDataAvailable(oldState)
          ? { data: oldState.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await schemasGetAll({
        params: {
          active: showActive,
          query: queryParam || undefined,
        },
        signal,
      });

      if (response.isSuccessful) {
        setSchemas({ data: response.data.schemas, status: "successful" });

        response.data.errors.forEach((zodError) => {
          processZodError(zodError).forEach((error) => void message.error(error));
        });
      } else {
        if (!isAbortedError(response.error)) {
          setSchemas({ error: response.error, status: "failed" });
        }
      }
    },
    [queryParam, showActive]
  );

  const tableContents: ColumnsType<Schema> = [
    {
      dataIndex: "schema",
      ellipsis: { showTitle: false },
      key: "schema",
      render: (name: string) => (
        <Tooltip placement="topLeft" title={name}>
          <Typography.Text strong>{name}</Typography.Text>
        </Tooltip>
      ),
      sorter: ({ schema: a }, { schema: b }) => a.localeCompare(b),
      title: FORM_LABEL.SCHEMA_TYPE,
    },
    {
      dataIndex: "createdAt",
      ellipsis: { showTitle: false },
      key: "createdAt",
      render: (createdAt: Date) => <Typography.Text>{formatDate(createdAt, true)}</Typography.Text>,
      sorter: ({ createdAt: a }, { createdAt: b }) => b.getTime() - a.getTime(),
      title: "Import date",
      width: 180,
    },
    {
      dataIndex: "id",
      key: "id",
      render: (schemaID: string) =>
        showActive ? (
          <Row justify="space-between">
            <Space>
              <Typography.Link onClick={() => openDetails(schemaID)}>Details</Typography.Link>
              <Link
                to={generatePath(ROUTES.issueClaim.path, {
                  schemaID,
                })}
              >
                Issue
              </Link>
            </Space>

            <SchemaRowDropdown id={schemaID} onAction={() => makeRequestAbortable(getSchemas)} />
          </Row>
        ) : (
          <Space size="middle">
            <Typography.Link onClick={() => openDetails(schemaID)}>Details</Typography.Link>
            <Typography.Link
              onClick={() => {
                handleActivate(schemaID);
              }}
            >
              Move to my schemas
            </Typography.Link>
          </Space>
        ),
      title: "Actions",
      width: showActive ? 180 : 250,
    },
  ];

  const openDetails = (schemaID: string) => {
    const params = new URLSearchParams(searchParams);

    params.set(SCHEMA_ID_SEARCH_PARAM, schemaID);
    setSearchParams(params);
  };

  const handleActivate = (id: string) => {
    void schemasUpdate({
      payload: { active: true },
      schemaID: id,
    }).then((isUpdated) => {
      if (isUpdated.isSuccessful) {
        void message.success("Claim schema moved to active.");
        makeRequestAbortable(getSchemas);
      }
    });
  };

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

  useEffect(() => {
    const { aborter } = makeRequestAbortable(getSchemas);

    return aborter;
  }, [getSchemas]);

  const schemaList = isAsyncTaskDataAvailable(schemas) ? schemas.data : [];

  return (
    <>
      <TableCard
        defaultContents={
          showActive ? (
            <>
              <Avatar className="avatar-color-cyan" icon={<IconCertificate />} size={48} />

              <Typography.Text strong>No schemas</Typography.Text>
              <Typography.Text type="secondary">
                Imported schemas will be listed here.
              </Typography.Text>

              <Link to={ROUTES.importSchema.path}>
                <Button icon={<IconUpload />} type="primary">
                  Import schema
                </Button>
              </Link>
            </>
          ) : (
            <>
              <Avatar className="avatar-color-cyan" icon={<IconArchive />} size={48} />
              <Space align="center" direction="vertical">
                <Typography.Text strong>No archived schemas.</Typography.Text>
                <Typography.Text type="secondary">
                  Archived schemas will be listed here.
                </Typography.Text>
              </Space>
            </>
          )
        }
        isLoading={isAsyncTaskStarting(schemas)}
        onSearch={onSearch}
        query={queryParam}
        showDefaultContents={
          schemas.status === "successful" && schemaList.length === 0 && queryParam === null
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
            dataSource={schemaList}
            locale={{
              emptyText:
                schemas.status === "failed" ? (
                  <ErrorResult error={schemas.error.message} />
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
            <Space align="end" size="middle">
              <Card.Meta title={showActive ? "My schemas" : "Archived schemas"} />

              <Tag color="blue">{schemaList.length}</Tag>
            </Space>
          </Row>
        }
      />

      <SchemaDetails />
    </>
  );
}
