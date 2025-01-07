import {
  Avatar,
  Button,
  Card,
  Dropdown,
  Row,
  Space,
  Table,
  TableColumnsType,
  Tag,
  Tooltip,
  Typography,
} from "antd";
import { useCallback, useEffect, useState } from "react";
import { generatePath, useNavigate, useSearchParams } from "react-router-dom";
import IconCheckMark from "src/assets/icons/check.svg?react";
import IconCopy from "src/assets/icons/copy-01.svg?react";
import IconDots from "src/assets/icons/dots-vertical.svg?react";
import IconIssuers from "src/assets/icons/fingerprint-02.svg?react";
import IconInfoCircle from "src/assets/icons/info-circle.svg?react";
import IconPlus from "src/assets/icons/plus.svg?react";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { NoResults } from "src/components/shared/NoResults";
import { TableCard } from "src/components/shared/TableCard";
import { useIdentityContext } from "src/contexts/Identity";
import { Identity } from "src/domain";
import { ROUTES } from "src/routes";
import { isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";

import {
  DETAILS,
  DOTS_DROPDOWN_WIDTH,
  IDENTITY_ADD,
  QUERY_SEARCH_PARAM,
} from "src/utils/constants";
import { formatIdentifier } from "src/utils/forms";

export function IdentitiesTable({ handleAddIdentity }: { handleAddIdentity: () => void }) {
  const { identityList } = useIdentityContext();
  const navigate = useNavigate();

  const [filteredIdentifiers, setFilteredIdentifiers] = useState<Identity[]>(() =>
    isAsyncTaskDataAvailable(identityList) ? identityList.data : []
  );
  const [searchParams, setSearchParams] = useSearchParams();

  const queryParam = searchParams.get(QUERY_SEARCH_PARAM);

  useEffect(() => {
    if (isAsyncTaskDataAvailable(identityList)) {
      if (!queryParam) {
        setFilteredIdentifiers(identityList.data);
      } else {
        const filteredData = identityList.data.filter((issuer: Identity) =>
          Object.values(issuer).some((value: string | null) =>
            value?.toLowerCase().includes(queryParam.toLowerCase())
          )
        );
        setFilteredIdentifiers(filteredData);
      }
    }
  }, [queryParam, identityList]);

  const onSearch = useCallback(
    (query: string) => {
      setSearchParams((previousParams) => {
        const previousQuery = previousParams.get(QUERY_SEARCH_PARAM);
        const params = new URLSearchParams(previousParams);

        if (query === "") {
          params.delete(QUERY_SEARCH_PARAM);
        } else if (previousQuery !== query) {
          params.set(QUERY_SEARCH_PARAM, query);
        }

        return params;
      });
    },
    [setSearchParams]
  );

  const tableColumns: TableColumnsType<Identity> = [
    {
      dataIndex: "displayName",
      key: "displayName",
      render: (displayName: Identity["displayName"], identity: Identity) => (
        <Typography.Link
          onClick={() =>
            navigate(
              generatePath(ROUTES.identityDetails.path, {
                identityID: identity.identifier,
              })
            )
          }
          strong
        >
          {displayName || formatIdentifier(identity.identifier, { short: true })}
        </Typography.Link>
      ),
      sorter: (a, b) =>
        (a.displayName || a.identifier).localeCompare(b.displayName || b.identifier),
      title: "Name",
    },
    {
      dataIndex: "identifier",
      key: "identifier",
      render: (identifier: Identity["identifier"]) => (
        <Tooltip title={identifier}>
          <Typography.Text
            copyable={{
              icon: [<IconCopy key={0} />, <IconCheckMark key={1} />],
            }}
            ellipsis={{
              suffix: identifier.slice(-5),
            }}
          >
            {identifier}
          </Typography.Text>
        </Tooltip>
      ),
      sorter: ({ identifier: a }, { identifier: b }) => a.localeCompare(b),
      title: "DID",
    },
    {
      dataIndex: "credentialStatusType",
      key: "credentialStatusType",
      render: (credentialStatusType: Identity["credentialStatusType"]) => (
        <Typography.Text>{credentialStatusType}</Typography.Text>
      ),
      sorter: ({ identifier: a }, { identifier: b }) => a.localeCompare(b),
      title: "Credential status",
    },
    {
      dataIndex: "blockchain",
      key: "blockchain",
      render: (blockchain: Identity["blockchain"]) => (
        <Typography.Text>{blockchain}</Typography.Text>
      ),
      sorter: ({ blockchain: a }, { blockchain: b }) => a.localeCompare(b),
      title: "Blockchain",
    },
    {
      dataIndex: "network",
      key: "network",
      render: (network: Identity["network"]) => <Typography.Text>{network}</Typography.Text>,
      sorter: ({ network: a }, { network: b }) => a.localeCompare(b),
      title: "Network",
    },
    {
      dataIndex: "identifier",
      key: "identifier",
      render: (identifier: Identity["identifier"]) => (
        <Dropdown
          menu={{
            items: [
              {
                icon: <IconInfoCircle />,
                key: "details",
                label: DETAILS,
                onClick: () =>
                  navigate(
                    generatePath(ROUTES.identityDetails.path, {
                      identityID: identifier,
                    })
                  ),
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

  const addButton = (
    <Button icon={<IconPlus />} onClick={handleAddIdentity} type="primary">
      {IDENTITY_ADD}
    </Button>
  );

  return (
    <TableCard
      defaultContents={
        <>
          <Avatar className="avatar-color-icon" icon={<IconIssuers />} size={48} />

          <Typography.Text strong>No issuers</Typography.Text>

          <Typography.Text type="secondary">
            Add a new issuer to get the required credential
          </Typography.Text>

          {addButton}
        </>
      }
      isLoading={isAsyncTaskStarting(identityList)}
      onSearch={onSearch}
      query={queryParam}
      searchPlaceholder="Search identity name, DID, etc..."
      showDefaultContents={
        identityList.status === "successful" &&
        filteredIdentifiers.length === 0 &&
        queryParam === null
      }
      table={
        <Table
          columns={tableColumns}
          dataSource={filteredIdentifiers}
          locale={{
            emptyText:
              identityList.status === "failed" ? (
                <ErrorResult error={identityList.error.message} />
              ) : (
                <NoResults searchQuery={queryParam} />
              ),
          }}
          pagination={false}
          rowKey="identifier"
          showSorterTooltip
          sortDirections={["ascend", "descend"]}
        />
      }
      title={
        <Row justify="space-between">
          <Space size="middle">
            <Card.Meta title="Identities" />
            <Tag>{filteredIdentifiers.length}</Tag>
          </Space>
        </Row>
      }
    />
  );
}
