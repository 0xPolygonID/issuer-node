import {
  Avatar,
  Button,
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
import { Link, generatePath, useSearchParams } from "react-router-dom";
import { z } from "zod";

import { Claim, ClaimAttribute, claimUpdate, claimsGetAll } from "src/adapters/api/claims";
import { Schema } from "src/adapters/api/schemas";
import { ReactComponent as IconLinkExternal } from "src/assets/icons/link-external-01.svg";
import { ErrorResult } from "src/components/schemas/ErrorResult";
import { IssuedClaimDetails } from "src/components/schemas/IssuedClaimDetails";
import { NoResults } from "src/components/schemas/NoResults";
import { TableCard } from "src/components/schemas/TableCard";
import { useAuthContext } from "src/hooks/useAuthContext";
import { ROUTES } from "src/routes";
import { APIError, processZodError, stringBoolean } from "src/utils/adapters";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import {
  CLAIM_ID_SEARCH_PARAM,
  FORM_LABEL,
  QUERY_SEARCH_PARAM,
  SCHEMAS_TABS,
  SCHEMA_ID_SEARCH_PARAM,
  VALID_SEARCH_PARAM,
} from "src/utils/constants";
import { formatDate } from "src/utils/forms";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/types";

export function IssuedClaims() {
  const [claims, setClaims] = useState<AsyncTask<Claim[], APIError>>({
    status: "pending",
  });
  const [isClaimUpdating, setClaimUpdating] = useState<Record<string, boolean>>({});
  const { account, authToken } = useAuthContext();
  const [searchParams, setSearchParams] = useSearchParams();
  const validParam = searchParams.get(VALID_SEARCH_PARAM);
  const parsedValidParam = stringBoolean.safeParse(validParam);
  const queryParam = searchParams.get(QUERY_SEARCH_PARAM);
  const [searchQuery, setSearchQuery] = useState<string | null>(queryParam);
  const showValid = parsedValidParam.success ? parsedValidParam.data : true;

  const getClaims = useCallback(
    async (signal: AbortSignal) => {
      if (authToken && account?.organization) {
        setClaims((previousClaims) =>
          isAsyncTaskDataAvailable(previousClaims)
            ? { data: previousClaims.data, status: "reloading" }
            : { status: "loading" }
        );

        const response = await claimsGetAll({
          issuerID: account.organization,
          params: {
            query: queryParam || undefined,
            valid: showValid,
          },
          signal,
          token: authToken,
        });

        if (response.isSuccessful) {
          setClaims({ data: response.data.claims, status: "successful" });

          response.data.errors.forEach((zodError) => {
            processZodError(zodError).forEach((error) => void message.error(error));
          });
        } else {
          if (!isAbortedError(response.error)) {
            setClaims({ error: response.error, status: "failed" });
          }
        }
      }
    },
    [authToken, account, queryParam, showValid]
  );

  const handleShowValidChange = ({ target: { value } }: RadioChangeEvent) => {
    const parsedShowValid = z.boolean().safeParse(value);
    if (parsedShowValid.success) {
      const params = new URLSearchParams(searchParams);

      params.set(VALID_SEARCH_PARAM, parsedShowValid.data.toString());

      setClaims({ status: "pending" });
      setSearchParams(params);
      setSearchQuery(null);
    } else {
      processZodError(parsedShowValid.error).forEach((error) => void message.error(error));
    }
  };

  const updateClaimInState = (claim: Claim) => {
    setClaims((oldClaims) =>
      isAsyncTaskDataAvailable(oldClaims)
        ? {
            data: oldClaims.data.map((currentClaim: Claim) =>
              currentClaim.id === claim.id ? claim : currentClaim
            ),
            status: "successful",
          }
        : oldClaims
    );
  };

  const toggleClaimActive = (active: boolean, { id: claimID, schemaTemplate }: Claim) => {
    if (authToken && account?.organization) {
      setClaimUpdating((currentClaimsUpdating) => {
        return { ...currentClaimsUpdating, [claimID]: true };
      });

      void claimUpdate({
        claimID,
        issuerID: account.organization,
        payload: { active },
        schemaID: schemaTemplate.id,
        token: authToken,
      }).then((response) => {
        if (response.isSuccessful) {
          updateClaimInState(response.data);

          void message.success(`Link ${active ? "activated" : "deactivated"}`);
        } else {
          void message.error(response.error.message);
        }

        setClaimUpdating((currentClaimsUpdating) => {
          return { ...currentClaimsUpdating, [claimID]: false };
        });
      });
    }
  };

  useEffect(() => {
    const { aborter } = makeRequestAbortable(getClaims);

    return aborter;
  }, [getClaims]);

  useEffect(() => {
    setSearchParams((oldSearchParams) => {
      const oldQuery = oldSearchParams.get(QUERY_SEARCH_PARAM);
      const params = new URLSearchParams(oldSearchParams);

      if (!searchQuery) {
        params.delete(QUERY_SEARCH_PARAM);

        return params;
      } else if (oldQuery !== searchQuery) {
        params.set(QUERY_SEARCH_PARAM, searchQuery);

        return params;
      } else {
        return oldSearchParams;
      }
    });
  }, [searchQuery, setSearchParams]);

  const TABLE_CONTENTS: ColumnsType<Claim> = [
    {
      dataIndex: "active",
      ellipsis: true,
      key: "active",
      render: (active: boolean, claim: Claim) => (
        <Switch
          checked={claim.valid && active}
          disabled={!claim.valid}
          loading={isClaimUpdating[claim.id]}
          onClick={(isActive, event) => {
            event.stopPropagation();
            toggleClaimActive(isActive, claim);
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
      title: FORM_LABEL.SCHEMA_NAME,
    },
    {
      dataIndex: "attributeValues",
      ellipsis: true,
      key: "attributeValues",
      render: (attributes: ClaimAttribute[]) => {
        const names = attributes.map(({ attributeKey }) => attributeKey).join(", ");

        return (
          <Tooltip placement="topLeft" title={names}>
            <Typography.Text>{names}</Typography.Text>
          </Tooltip>
        );
      },
      title: FORM_LABEL.ATTRIBUTES,
    },
    {
      dataIndex: "claimLinkExpiration",
      ellipsis: true,
      key: "claimLinkExpiration",
      render: (date: Date | null) => (
        <Typography.Text>{date ? formatDate(date, true) : "-"}</Typography.Text>
      ),
      title: FORM_LABEL.LINK_VALIDITY,
    },
    {
      dataIndex: "issuedClaims",
      ellipsis: true,
      key: "issuedClaims",
      render: (issued: number | null, claim) => {
        const limit = claim.limitedClaims;
        const left = issued !== null && limit !== null ? limit - issued : null;
        const value = limit !== null && left !== null ? `${left} of ${limit}` : "-";

        return <Typography.Text>{value}</Typography.Text>;
      },
      title: FORM_LABEL.CLAIM_AVAILABILITY,
    },
  ];

  const claimsList = isAsyncTaskDataAvailable(claims) ? claims.data : [];

  return (
    <>
      <TableCard
        defaultContents={
          showValid ? (
            <>
              <Avatar className="avatar-color-cyan" icon={<IconLinkExternal />} size={48} />

              <Typography.Text strong>No expired shared links.</Typography.Text>
            </>
          ) : (
            <>
              <Avatar className="avatar-color-cyan" icon={<IconLinkExternal />} size={48} />

              <Typography.Text strong>No shared links.</Typography.Text>

              <Link
                to={generatePath(ROUTES.schemas.path, {
                  tabID: SCHEMAS_TABS[0].tabID,
                })}
              >
                <Button>{FORM_LABEL.MY_SCHEMAS}</Button>
              </Link>
            </>
          )
        }
        isLoading={isAsyncTaskStarting(claims)}
        onSearch={setSearchQuery}
        query={queryParam}
        showDefaultContents={
          claims.status === "successful" && claimsList.length === 0 && searchQuery === null
        }
        table={
          <Table
            columns={TABLE_CONTENTS.map(({ title, ...column }) => ({
              title: (
                <Typography.Text type="secondary">
                  <>{title}</>
                </Typography.Text>
              ),
              ...column,
            }))}
            dataSource={claimsList}
            locale={{
              emptyText:
                claims.status === "failed" ? (
                  <ErrorResult error={claims.error.message} />
                ) : (
                  <NoResults searchQuery={queryParam} />
                ),
            }}
            onRow={({ id, schemaTemplate }: Claim) => ({
              onClick: () => {
                const params = new URLSearchParams(searchParams);

                params.set(CLAIM_ID_SEARCH_PARAM, id);
                params.set(SCHEMA_ID_SEARCH_PARAM, schemaTemplate.id);
                setSearchParams(params);
              },
            })}
            pagination={false}
            rowClassName="pointer"
            rowKey="id"
            showSorterTooltip
            sortDirections={["ascend", "descend"]}
          />
        }
        title={
          <Row justify="space-between">
            <Space size="middle">
              <Card.Meta title="Claim links" />

              <Tag color="blue">
                {`${claimsList.length} ${claimsList.length === 1 ? "link" : "links"}`}
              </Tag>
            </Space>

            <Radio.Group onChange={handleShowValidChange} value={showValid}>
              <Radio.Button value={true}>Valid</Radio.Button>

              <Radio.Button value={false}>Expired</Radio.Button>
            </Radio.Group>
          </Row>
        }
      />

      <IssuedClaimDetails />
    </>
  );
}
