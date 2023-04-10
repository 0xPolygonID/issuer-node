import {
  Avatar,
  Button,
  Card,
  Dropdown,
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
import { generatePath, useNavigate, useSearchParams } from "react-router-dom";

import { APIError } from "src/adapters/api";
import { getLinks, linkStatusParser, linkUpdate } from "src/adapters/api/credentials";
import { ReactComponent as IconCreditCardPlus } from "src/assets/icons/credit-card-plus.svg";
import { ReactComponent as IconDots } from "src/assets/icons/dots-vertical.svg";
import { ReactComponent as IconInfoCircle } from "src/assets/icons/info-circle.svg";
import { ReactComponent as IconLink } from "src/assets/icons/link-03.svg";
import { ReactComponent as IconTrash } from "src/assets/icons/trash-01.svg";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { NoResults } from "src/components/shared/NoResults";
import { TableCard } from "src/components/shared/TableCard";
import { useEnvContext } from "src/contexts/env";
import { Link } from "src/domain/credential";
import { ROUTES } from "src/routes";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import {
  ACCESSIBLE_UNTIL,
  LINKS,
  QUERY_SEARCH_PARAM,
  STATUS_SEARCH_PARAM,
} from "src/utils/constants";
import { processZodError } from "src/utils/error";
import { formatDate } from "src/utils/forms";

export function LinkTable() {
  const env = useEnvContext();

  const navigate = useNavigate();

  const [links, setLinks] = useState<AsyncTask<Link[], APIError>>({
    status: "pending",
  });
  const [isLinkUpdating, setLinkUpdating] = useState<Record<string, boolean>>({});

  const [searchParams, setSearchParams] = useSearchParams();

  const linksList = isAsyncTaskDataAvailable(links) ? links.data : [];
  const statusParam = searchParams.get(STATUS_SEARCH_PARAM);
  const queryParam = searchParams.get(QUERY_SEARCH_PARAM);
  const parsedStatusParam = linkStatusParser.safeParse(statusParam);

  const status = parsedStatusParam.success ? parsedStatusParam.data : "all";

  const tableColumns: ColumnsType<Link> = [
    {
      dataIndex: "active",
      ellipsis: true,
      key: "active",
      render: (active: Link["active"], link: Link) => (
        <Switch
          checked={active}
          disabled={link.status === "exceeded"}
          loading={isLinkUpdating[link.id]}
          onClick={(isActive, event) => {
            event.stopPropagation();
            toggleCredentialActive(isActive, link);
          }}
          size="small"
        />
      ),
      sorter: ({ active: a }, { active: b }) => (a === b ? 0 : a ? 1 : -1),
      title: "Active",
      width: 100,
    },
    {
      dataIndex: "schemaType",
      ellipsis: true,
      key: "schemaType",
      render: (schemaType: Link["schemaType"]) => (
        <Tooltip placement="topLeft" title={schemaType}>
          <Typography.Text strong>{schemaType}</Typography.Text>
        </Tooltip>
      ),
      sorter: ({ schemaType: a }, { schemaType: b }) => a.localeCompare(b),
      title: "Credential",
    },
    {
      dataIndex: "expiration",
      ellipsis: true,
      key: "expiration",
      render: (expiration: Link["expiration"]) => (
        <Typography.Text>{expiration ? formatDate(expiration, true) : "Unlimited"}</Typography.Text>
      ),
      title: ACCESSIBLE_UNTIL,
    },
    {
      dataIndex: "issuedClaims",
      ellipsis: true,
      key: "issuedClaims",
      render: (issuedClaims: Link["issuedClaims"]) => {
        const value = issuedClaims ? issuedClaims : 0;

        return <Typography.Text>{value}</Typography.Text>;
      },
      title: "Credentials issued",
    },
    {
      dataIndex: "maxIssuance",
      ellipsis: true,
      key: "maxIssuance",
      render: (maxIssuance: Link["maxIssuance"]) => {
        const value = maxIssuance ? maxIssuance : "Unlimited";

        return <Typography.Text>{value}</Typography.Text>;
      },
      title: "Maximum issuance",
    },
    {
      dataIndex: "status",
      key: "status",
      render: (status: Link["status"]) => (
        <Row justify="space-between">
          {(() => {
            switch (status) {
              case "active": {
                return <Tag color="success">Active</Tag>;
              }
              case "inactive": {
                return <Tag>Inactive</Tag>;
              }
              case "exceeded": {
                return <Tag color="error">Exceeded</Tag>;
              }
            }
          })()}
          <Dropdown
            menu={{
              items: [
                {
                  icon: <IconInfoCircle />,
                  key: "details",
                  label: "Details",
                },
                {
                  key: "divider",
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
        </Row>
      ),
      title: "Status",
      width: 140,
    },
  ];

  const fetchLinks = useCallback(
    async (signal: AbortSignal) => {
      setLinks((previousLinks) =>
        isAsyncTaskDataAvailable(previousLinks)
          ? { data: previousLinks.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await getLinks({
        env,
        params: {
          query: queryParam || undefined,
          status: status,
        },
        signal,
      });

      if (response.isSuccessful) {
        setLinks({ data: response.data, status: "successful" });
      } else {
        if (!isAbortedError(response.error)) {
          setLinks({ error: response.error, status: "failed" });
        }
      }
    },
    [env, queryParam, status]
  );

  const handleStatusChange = ({ target: { value } }: RadioChangeEvent) => {
    const parsedLinkValue = linkStatusParser.safeParse(value);
    if (parsedLinkValue.success) {
      const params = new URLSearchParams(searchParams);
      if (parsedLinkValue.data !== "all") {
        params.set(STATUS_SEARCH_PARAM, parsedLinkValue.data);
      } else {
        params.delete(STATUS_SEARCH_PARAM);
      }

      setLinks({ status: "pending" });
      setSearchParams(params);
    } else {
      processZodError(parsedLinkValue.error).forEach((error) => void message.error(error));
    }
  };

  const updateCredentialInState = (link: Link) => {
    setLinks((oldLinks) =>
      isAsyncTaskDataAvailable(oldLinks)
        ? {
            data: oldLinks.data.map((currentLink: Link) =>
              currentLink.id === link.id ? link : currentLink
            ),
            status: "successful",
          }
        : oldLinks
    );
  };

  const toggleCredentialActive = (active: boolean, { id }: Link) => {
    setLinkUpdating((currentLinksUpdating) => {
      return { ...currentLinksUpdating, [id]: true };
    });

    void linkUpdate({
      env,
      id,
      payload: { active },
    }).then((response) => {
      if (response.isSuccessful) {
        updateCredentialInState(response.data);

        void message.success(`Link ${active ? "activated" : "deactivated"}`);
      } else {
        void message.error(response.error.message);
      }

      setLinkUpdating((currentLinksUpdating) => {
        return { ...currentLinksUpdating, [id]: false };
      });
    });
  };

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchLinks);

    return aborter;
  }, [fetchLinks]);

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

  const showDefaultContent =
    links.status === "successful" && linksList.length === 0 && queryParam === null;

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

            {status === "all" && (
              <Button
                icon={<IconCreditCardPlus />}
                onClick={() => navigate(generatePath(ROUTES.issueCredential.path))}
                type="primary"
              >
                Issue credential
              </Button>
            )}
          </>
        }
        isLoading={isAsyncTaskStarting(links)}
        onSearch={onSearch}
        query={queryParam}
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
            dataSource={linksList}
            locale={{
              emptyText:
                links.status === "failed" ? (
                  <ErrorResult error={links.error.message} />
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

              <Tag color="blue">{linksList.length}</Tag>
            </Space>

            {(!showDefaultContent || status !== "all") && (
              <Radio.Group onChange={handleStatusChange} value={status}>
                <Radio.Button value="all">All</Radio.Button>

                <Radio.Button value="active">Active</Radio.Button>

                <Radio.Button value="inactive">Inactive</Radio.Button>

                <Radio.Button value="exceeded">Exceeded</Radio.Button>
              </Radio.Group>
            )}
          </Row>
        }
      />
    </>
  );
}
