import {
  Avatar,
  Button,
  Card,
  Dropdown,
  Flex,
  Row,
  Space,
  Table,
  TableColumnsType,
  Tag,
  Typography,
} from "antd";
import { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

import { getCredentialsByIDs } from "src/adapters/api/credentials";
import IconCreditCardRefresh from "src/assets/icons/credit-card-refresh.svg?react";
import IconDots from "src/assets/icons/dots-vertical.svg?react";
import IconPlus from "src/assets/icons/plus.svg?react";
import IconClose from "src/assets/icons/x.svg?react";
import { CredentialRevokeModal } from "src/components/shared/CredentialRevokeModal";
import { TableCard } from "src/components/shared/TableCard";

import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { AppError, Credential } from "src/domain";
import { ROUTES } from "src/routes";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { DOTS_DROPDOWN_WIDTH, ISSUED, ISSUE_DATE, REVOCATION, REVOKE } from "src/utils/constants";
import { notifyParseErrors } from "src/utils/error";
import { formatDate } from "src/utils/forms";

export function IdentityAuthCredentials({ IDs }: { IDs: Array<string> }) {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();
  const navigate = useNavigate();

  const [credentials, setCredentials] = useState<AsyncTask<Credential[], AppError>>({
    status: "pending",
  });
  const [credentialToRevoke, setCredentialToRevoke] = useState<Credential>();

  const fetchAuthCredentials = useCallback(
    async (signal?: AbortSignal) => {
      setCredentials((previousCredentials) =>
        isAsyncTaskDataAvailable(previousCredentials)
          ? { data: previousCredentials.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await getCredentialsByIDs({
        env,
        identifier,
        IDs,
        signal,
      });
      if (response.success) {
        setCredentials({
          data: response.data.successful,
          status: "successful",
        });

        notifyParseErrors(response.data.failed);
      } else {
        if (!isAbortedError(response.error)) {
          setCredentials({ error: response.error, status: "failed" });
        }
      }
    },
    [env, identifier, IDs]
  );

  const tableColumns: TableColumnsType<Credential> = [
    {
      dataIndex: "schemaType",
      ellipsis: { showTitle: false },
      key: "schemaType",
      render: (schemaType: Credential["schemaType"]) => (
        <Typography.Text strong>{schemaType}</Typography.Text>
      ),
      sorter: {
        compare: ({ schemaType: a }, { schemaType: b }) => (a && b ? a.localeCompare(b) : 0),
        multiple: 1,
      },
      title: "Type",
    },
    {
      dataIndex: "createdAt",
      key: "createdAt",
      render: (issuanceDate: Credential["issuanceDate"]) => (
        <Typography.Text>{formatDate(issuanceDate)}</Typography.Text>
      ),
      sorter: ({ issuanceDate: a }, { issuanceDate: b }) => b.getTime() - a.getTime(),
      title: ISSUE_DATE,
    },

    {
      dataIndex: "credentialSubject",
      ellipsis: { showTitle: false },
      key: "credentialSubject",
      render: (credentialSubject: Credential["credentialSubject"]) => (
        <Typography.Text>
          {typeof credentialSubject.x === "string" ? credentialSubject.x : "-"}
        </Typography.Text>
      ),

      title: "Credential Subject x",
    },
    {
      dataIndex: "credentialSubject",
      ellipsis: { showTitle: false },
      key: "credentialSubject",
      render: (credentialSubject: Credential["credentialSubject"]) => (
        <Typography.Text>
          {typeof credentialSubject.y === "string" ? credentialSubject.y : "-"}
        </Typography.Text>
      ),
      title: "Credential Subject y",
    },
    {
      dataIndex: "credentialStatus",
      ellipsis: { showTitle: false },
      key: "credentialStatus",
      render: (credentialStatus: Credential["credentialStatus"]) => (
        <Typography.Text>{credentialStatus.revocationNonce}</Typography.Text>
      ),
      title: "Revocation nonce",
    },
    {
      dataIndex: "credentialStatus",
      ellipsis: { showTitle: false },
      key: "credentialStatus",
      render: (credentialStatus: Credential["credentialStatus"]) => (
        <Typography.Text>{credentialStatus.type}</Typography.Text>
      ),
      title: "Revocation status type",
    },
    {
      dataIndex: "revoked",
      key: "revoked",
      render: (revoked: Credential["revoked"]) => (
        <Typography.Text>{revoked ? "Revoked" : "-"}</Typography.Text>
      ),
      responsive: ["sm"],
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
                danger: true,
                disabled: credential.revoked,
                icon: <IconClose />,
                key: "revoke",
                label: REVOKE,
                onClick: () => setCredentialToRevoke(credential),
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

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchAuthCredentials);

    return aborter;
  }, [fetchAuthCredentials]);

  const credentialsList = isAsyncTaskDataAvailable(credentials) ? credentials.data : [];
  const showDefaultContent = credentials.status === "successful" && credentialsList.length === 0;

  return (
    <Card
      title={
        <Flex align="center" justify="space-between" style={{ padding: 12 }}>
          <Typography.Text style={{ fontSize: 18 }}>Auth Credentials</Typography.Text>

          <Button
            icon={<IconPlus />}
            onClick={() => navigate(ROUTES.createAuthCredential.path)}
            type="primary"
          >
            Create auth credential
          </Button>
        </Flex>
      }
    >
      <TableCard
        defaultContents={
          <>
            <Avatar className="avatar-color-icon" icon={<IconCreditCardRefresh />} size={48} />

            <Typography.Text strong>No auth credentials</Typography.Text>

            <Typography.Text type="secondary">
              Auth credentials will be listed here.
            </Typography.Text>
          </>
        }
        isLoading={isAsyncTaskStarting(credentials)}
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
            loading={credentials.status === "reloading"}
            pagination={false}
            rowKey="id"
            showSorterTooltip
            sortDirections={["ascend", "descend"]}
          />
        }
        title={
          <Row gutter={[0, 8]} justify="space-between">
            <Space size="middle">
              <Card.Meta title={ISSUED} />

              <Tag>{credentialsList.length}</Tag>
            </Space>
          </Row>
        }
      />
      {credentialToRevoke && (
        <CredentialRevokeModal
          credential={credentialToRevoke}
          onClose={() => setCredentialToRevoke(undefined)}
          onRevoke={() => void fetchAuthCredentials()}
        />
      )}
    </Card>
  );
}
