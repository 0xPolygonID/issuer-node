import {
  Avatar,
  Button,
  Card,
  Dropdown,
  Row,
  Space,
  Table,
  TableColumnsType,
  Typography,
} from "antd";
import { useCallback, useEffect, useState } from "react";
import { generatePath, useNavigate, useSearchParams } from "react-router-dom";
import IconIssuers from "src/assets/icons/building-08.svg?react";
import IconDots from "src/assets/icons/dots-vertical.svg?react";
import IconInfoCircle from "src/assets/icons/info-circle.svg?react";
import IconPlus from "src/assets/icons/plus.svg?react";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { NoResults } from "src/components/shared/NoResults";
import { TableCard } from "src/components/shared/TableCard";
import { useIssuerContext } from "src/contexts/Issuer";
import { Issuer } from "src/domain";
import { ROUTES } from "src/routes";
import { isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";

import { DETAILS, DOTS_DROPDOWN_WIDTH, ISSUER_ADD, QUERY_SEARCH_PARAM } from "src/utils/constants";
import { formatIdentifier } from "src/utils/forms";

export function IssuersTable({ handleAddIssuer }: { handleAddIssuer: () => void }) {
  const { issuersList } = useIssuerContext();
  const navigate = useNavigate();

  const [filteredIdentifiers, setFilteredIdentifiers] = useState<Issuer[]>(() =>
    isAsyncTaskDataAvailable(issuersList) ? issuersList.data : []
  );
  const [searchParams, setSearchParams] = useSearchParams();

  const queryParam = searchParams.get(QUERY_SEARCH_PARAM);

  useEffect(() => {
    if (isAsyncTaskDataAvailable(issuersList)) {
      if (!queryParam) {
        setFilteredIdentifiers(issuersList.data);
      } else {
        const filteredData = issuersList.data.filter((issuer: Issuer) =>
          Object.values(issuer).some((value: string) => value.includes(queryParam))
        );
        setFilteredIdentifiers(filteredData);
      }
    }
  }, [queryParam, issuersList]);

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

  const tableColumns: TableColumnsType<Issuer> = [
    {
      dataIndex: "displayName",
      key: "displayName",
      render: (displayName: Issuer["displayName"]) => (
        <Typography.Text strong>{displayName}</Typography.Text>
      ),
      sorter: ({ displayName: a }, { displayName: b }) => a.localeCompare(b),
      title: "Name",
    },
    {
      dataIndex: "identifier",
      key: "identifier",
      render: (identifier: Issuer["identifier"]) => (
        <Typography.Text strong>{formatIdentifier(identifier)}</Typography.Text>
      ),
      sorter: ({ identifier: a }, { identifier: b }) => a.localeCompare(b),
      title: "DID",
    },
    {
      dataIndex: "blockchain",
      key: "blockchain",
      render: (blockchain: Issuer["blockchain"]) => (
        <Typography.Text strong>{blockchain}</Typography.Text>
      ),
      sorter: ({ blockchain: a }, { blockchain: b }) => a.localeCompare(b),
      title: "Blockchain",
    },
    {
      dataIndex: "network",
      key: "network",
      render: (network: Issuer["network"]) => <Typography.Text strong>{network}</Typography.Text>,
      sorter: ({ network: a }, { network: b }) => a.localeCompare(b),
      title: "Network",
    },
    {
      dataIndex: "identifier",
      key: "identifier",
      render: (identifier: Issuer["identifier"]) => (
        <Dropdown
          menu={{
            items: [
              {
                icon: <IconInfoCircle />,
                key: "details",
                label: DETAILS,
                onClick: () =>
                  navigate(
                    generatePath(ROUTES.issuerDetails.path, {
                      issuerID: identifier,
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
    <Button icon={<IconPlus />} onClick={handleAddIssuer} type="primary">
      {ISSUER_ADD}
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
      isLoading={isAsyncTaskStarting(issuersList)}
      onSearch={onSearch}
      query={queryParam}
      searchPlaceholder="Search Issuer"
      showDefaultContents={
        issuersList.status === "successful" &&
        filteredIdentifiers.length === 0 &&
        queryParam === null
      }
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
          dataSource={filteredIdentifiers}
          locale={{
            emptyText:
              issuersList.status === "failed" ? (
                <ErrorResult error={issuersList.error.message} />
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
          </Space>
        </Row>
      }
    />
  );
}
