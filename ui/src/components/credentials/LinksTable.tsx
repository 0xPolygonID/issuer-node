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
import dayjs from "dayjs";
import { useCallback, useEffect, useState } from "react";
import { generatePath, useNavigate, useSearchParams } from "react-router-dom";

import { APIError } from "src/adapters/api";
import { getLinks, linkStatusParser, updateLink } from "src/adapters/api/credentials";
import { ReactComponent as IconCreditCardPlus } from "src/assets/icons/credit-card-plus.svg";
import { ReactComponent as IconDots } from "src/assets/icons/dots-vertical.svg";
import { ReactComponent as IconInfoCircle } from "src/assets/icons/info-circle.svg";
import { ReactComponent as IconLink } from "src/assets/icons/link-03.svg";
import { ReactComponent as IconTrash } from "src/assets/icons/trash-01.svg";
import { LinkDeleteModal } from "src/components/credentials/LinkDeleteModal";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { NoResults } from "src/components/shared/NoResults";
import { TableCard } from "src/components/shared/TableCard";
import { useEnvContext } from "src/contexts/Env";
import { Link } from "src/domain";
import { ROUTES } from "src/routes";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import {
  ACCESSIBLE_UNTIL,
  LINKS,
  QUERY_SEARCH_PARAM,
  STATUS,
  STATUS_SEARCH_PARAM,
} from "src/utils/constants";
import { formatDate } from "src/utils/forms";

export function LinksTable() {
  const env = useEnvContext();

  const navigate = useNavigate();

  const [links, setLinks] = useState<AsyncTask<Link[], APIError>>({
    status: "pending",
  });
  const [isLinkUpdating, setLinkUpdating] = useState<Record<string, boolean>>({});
  const [linkToDelete, setLinkToDelete] = useState<string>();

  const [searchParams, setSearchParams] = useSearchParams();

  const linksList = isAsyncTaskDataAvailable(links) ? links.data : [];
  const statusParam = searchParams.get(STATUS_SEARCH_PARAM);
  const queryParam = searchParams.get(QUERY_SEARCH_PARAM);
  const parsedStatusParam = linkStatusParser.safeParse(statusParam);
  const showDefaultContent =
    links.status === "successful" && linksList.length === 0 && queryParam === null;

  const status = parsedStatusParam.success ? parsedStatusParam.data : undefined;

  const tableColumns: ColumnsType<Link> = [
    {
      dataIndex: "active",
      ellipsis: true,
      key: "active",
      render: (active: Link["active"], link: Link) => (
        <Switch
          checked={active && link.status !== "exceeded"}
          disabled={link.status === "exceeded"}
          loading={isLinkUpdating[link.id]}
          onClick={(isActive) => {
            toggleLinkActive(isActive, link.id);
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
        <Typography.Text>{expiration ? formatDate(expiration) : "Unlimited"}</Typography.Text>
      ),
      sorter: ({ expiration: a }, { expiration: b }) => {
        if (a && b) {
          return dayjs(a).unix() - dayjs(b).unix();
        } else if (a) {
          return 1;
        } else {
          return -1;
        }
      },
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
      sorter: ({ issuedClaims: a }, { issuedClaims: b }) => (a === b ? 0 : a ? 1 : -1),
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
      sorter: ({ maxIssuance: a }, { maxIssuance: b }) => {
        if (a && b) {
          return a - b;
        } else if (a) {
          return 1;
        } else {
          return -1;
        }
      },
      title: "Maximum issuance",
    },
    {
      dataIndex: "status",
      key: "status",
      render: (status: Link["status"], { id }: Link) => (
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
                  onClick: () => navigate(generatePath(ROUTES.linkDetails.path, { linkID: id })),
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
                  onClick: () => setLinkToDelete(id),
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
      sorter: ({ status: a }, { status: b }) => a.localeCompare(b),
      title: STATUS,
      width: 140,
    },
  ];

  const fetchLinks = useCallback(
    async (signal?: AbortSignal) => {
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
    const params = new URLSearchParams(searchParams);

    if (parsedLinkValue.success) {
      params.set(STATUS_SEARCH_PARAM, parsedLinkValue.data);
    } else {
      params.delete(STATUS_SEARCH_PARAM);
    }
    setSearchParams(params);
    setLinks({ status: "pending" });
  };

  const updateCredentialInState = (active: Link["active"], id: Link["id"]) => {
    setLinks((previousLinks) =>
      isAsyncTaskDataAvailable(previousLinks)
        ? {
            data: previousLinks.data.reduce((links: Link[], currentLink: Link) => {
              if (currentLink.id === id) {
                if (status === currentLink.status) {
                  return links;
                } else {
                  const linkStatusInverted: Link = {
                    ...currentLink,
                    active,
                    status: currentLink.status === "active" ? "inactive" : "active",
                  };

                  return [...links, linkStatusInverted];
                }
              } else {
                return [...links, currentLink];
              }
            }, []),
            status: "successful",
          }
        : previousLinks
    );
  };

  const toggleLinkActive = (active: boolean, id: Link["id"]) => {
    setLinkUpdating((currentLinksUpdating) => {
      return { ...currentLinksUpdating, [id]: true };
    });

    void updateLink({
      env,
      id,
      payload: { active },
    }).then((response) => {
      if (response.isSuccessful) {
        updateCredentialInState(active, id);

        void message.success(response.data);
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
        return previousParams;
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

            {status === undefined && (
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

            {(!showDefaultContent || status !== undefined) && (
              <Radio.Group onChange={handleStatusChange} value={status}>
                <Radio.Button value={undefined}>All</Radio.Button>

                <Radio.Button value="active">Active</Radio.Button>

                <Radio.Button value="inactive">Inactive</Radio.Button>

                <Radio.Button value="exceeded">Exceeded</Radio.Button>
              </Radio.Group>
            )}
          </Row>
        }
      />
      {linkToDelete && (
        <LinkDeleteModal
          id={linkToDelete}
          onClose={() => setLinkToDelete(undefined)}
          onDelete={() => void fetchLinks()}
        />
      )}
    </>
  );
}
