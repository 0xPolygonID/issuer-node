import {
  Avatar,
  Button,
  Card,
  Radio,
  RadioChangeEvent,
  Row,
  Space,
  Table,
  TableColumnsType,
  Typography,
} from "antd";
import { useCallback, useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { ErrorResult } from "../shared/ErrorResult";
import { NoResults } from "../shared/NoResults";
import { TableCard } from "../shared/TableCard";
import { identifierParser } from "src/adapters/api/issuers";
import IconIssuers from "src/assets/icons/building.svg?react";
import IconPlus from "src/assets/icons/plus.svg?react";
import { useIssuerContext } from "src/contexts/Issuer";
import { Issuer } from "src/domain/identifier";
import { isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";

import { QUERY_SEARCH_PARAM } from "src/utils/constants";

export function IssuersTable({ handleAddIssuer }: { handleAddIssuer: () => void }) {
  const { handleChange, identifier, issuersList } = useIssuerContext();
  const [filteredIdentifiers, setFilteredIdentifiers] = useState<Issuer[]>(() =>
    isAsyncTaskDataAvailable(issuersList) ? issuersList.data : []
  );
  const [searchParams, setSearchParams] = useSearchParams();

  const queryParam = searchParams.get(QUERY_SEARCH_PARAM);

  const handleIssuerChange = ({ target: { value } }: RadioChangeEvent) => {
    const parsedIdentifier = identifierParser.safeParse(value);

    if (parsedIdentifier.success) {
      handleChange(parsedIdentifier.data);
    } else {
      handleChange(null);
    }
  };

  useEffect(() => {
    if (isAsyncTaskDataAvailable(issuersList)) {
      if (queryParam === null) {
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
      align: "center",
      dataIndex: "identifier",
      key: "identifier",
      render: (identifier: Issuer["identifier"]) => <Radio.Button value={identifier} />,
    },
    {
      dataIndex: "identifier",
      key: "identifier",
      render: (identifier: Issuer["identifier"]) => (
        <Typography.Text strong>{identifier}</Typography.Text>
      ),
      sorter: ({ identifier: a }, { identifier: b }) => a.localeCompare(b),
      title: "Did",
    },
    {
      dataIndex: "authBJJCredentialStatus",
      key: "authBJJCredentialStatus",
      render: (credentialStatus: Issuer["authBJJCredentialStatus"]) => (
        <Typography.Text strong>{credentialStatus}</Typography.Text>
      ),
      sorter: ({ authBJJCredentialStatus: a }, { authBJJCredentialStatus: b }) =>
        a.localeCompare(b),

      title: "Credential Status",
    },
    {
      dataIndex: "network",
      key: "network",
      render: (network: Issuer["network"], { blockchain }: Issuer) => (
        <Typography.Text strong>
          {blockchain}-{network}
        </Typography.Text>
      ),
      sorter: (
        { blockchain: blockchainA, network: networkA },
        { blockchain: blockchainB, network: networkB }
      ) => `${blockchainA}-${networkA}`.localeCompare(`${blockchainB}-${networkB}`),
      title: "Network",
    },
  ];

  const addButton = (
    <Button icon={<IconPlus />} onClick={handleAddIssuer} type="primary">
      Add new issuer
    </Button>
  );

  return (
    <Radio.Group
      onChange={handleIssuerChange}
      optionType="default"
      size="small"
      style={{ width: "100%" }}
      value={identifier}
    >
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
        extraButton={addButton}
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
              <Card.Meta title="List of Issuers" />
            </Space>
          </Row>
        }
      />
    </Radio.Group>
  );
}
