import { PublicKey } from "@iden3/js-crypto";
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
  Tooltip,
  Typography,
} from "antd";
import { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

import { getAuthCredentialsByIDs } from "src/adapters/api/credentials";
import { notifyErrors } from "src/adapters/parsers";
import IconCheckMark from "src/assets/icons/check.svg?react";
import IconCopy from "src/assets/icons/copy-01.svg?react";
import IconCreditCardRefresh from "src/assets/icons/credit-card-refresh.svg?react";
import IconDots from "src/assets/icons/dots-vertical.svg?react";
import IconPlus from "src/assets/icons/plus.svg?react";
import IconClose from "src/assets/icons/x.svg?react";
import { CredentialRevokeModal } from "src/components/shared/CredentialRevokeModal";
import { TableCard } from "src/components/shared/TableCard";

import { useEnvContext } from "src/contexts/Env";
import { AppError, AuthCredential } from "src/domain";
import { ROUTES } from "src/routes";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { DOTS_DROPDOWN_WIDTH, ISSUE_DATE, REVOCATION, REVOKE } from "src/utils/constants";
import { formatDate } from "src/utils/forms";

export function IdentityAuthCredentials({
  identityID,
  IDs,
}: {
  IDs: Array<string>;
  identityID: string;
}) {
  const env = useEnvContext();
  const navigate = useNavigate();

  const [credentials, setCredentials] = useState<AsyncTask<AuthCredential[], AppError>>({
    status: "pending",
  });

  const [credentialToRevoke, setCredentialToRevoke] = useState<AuthCredential>();

  const fetchAuthCredentials = useCallback(
    async (signal?: AbortSignal) => {
      setCredentials((previousCredentials) =>
        isAsyncTaskDataAvailable(previousCredentials)
          ? { data: previousCredentials.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await getAuthCredentialsByIDs({
        env,
        identifier: identityID,
        IDs,
        signal,
      });
      if (response.success) {
        setCredentials({
          data: response.data.successful,
          status: "successful",
        });

        void notifyErrors(response.data.failed.filter((error) => error.type !== "cancel-error"));
      } else {
        if (!isAbortedError(response.error)) {
          setCredentials({ error: response.error, status: "failed" });
        }
      }
    },
    [env, identityID, IDs]
  );

  const tableColumns: TableColumnsType<AuthCredential> = [
    {
      dataIndex: "credentialSubject",
      ellipsis: { showTitle: false },
      key: "credentialSubject",
      render: (credentialSubject: AuthCredential["credentialSubject"]) => {
        const { x, y } = credentialSubject;
        const pKey = new PublicKey([x, y]);
        const pKeyHex = pKey.hex();
        return (
          <Tooltip title={pKeyHex}>
            <Typography.Text
              copyable={{
                icon: [<IconCopy key={0} />, <IconCheckMark key={1} />],
                text: pKeyHex,
              }}
              ellipsis={{
                suffix: pKeyHex.slice(-5),
              }}
            >
              {pKeyHex}
            </Typography.Text>
          </Tooltip>
        );
      },

      title: "Public key",
    },
    {
      dataIndex: "credentialStatus",
      ellipsis: { showTitle: false },
      key: "credentialStatus",
      render: (credentialStatus: AuthCredential["credentialStatus"]) => (
        <Typography.Text>{credentialStatus.revocationNonce}</Typography.Text>
      ),
      title: "Revocation nonce",
    },
    {
      dataIndex: "schemaType",
      ellipsis: { showTitle: false },
      key: "schemaType",
      render: (schemaType: AuthCredential["schemaType"]) => (
        <Typography.Text strong>{schemaType}</Typography.Text>
      ),
      sorter: {
        compare: ({ schemaType: a }, { schemaType: b }) => (a && b ? a.localeCompare(b) : 0),
        multiple: 1,
      },
      title: "Type",
    },
    {
      dataIndex: "credentialStatus",
      ellipsis: { showTitle: false },
      key: "credentialStatus",
      render: (credentialStatus: AuthCredential["credentialStatus"]) => (
        <Typography.Text>{credentialStatus.type}</Typography.Text>
      ),
      title: "Revocation status",
    },
    {
      dataIndex: "createdAt",
      key: "createdAt",
      render: (_, { issuanceDate }: AuthCredential) => (
        <Typography.Text>{formatDate(issuanceDate)}</Typography.Text>
      ),
      sorter: ({ issuanceDate: a }, { issuanceDate: b }) => b.getTime() - a.getTime(),
      title: ISSUE_DATE,
    },
    {
      dataIndex: "revoked",
      key: "revoked",
      render: (revoked: AuthCredential["revoked"]) => (
        <Typography.Text>{revoked ? "Revoked" : "-"}</Typography.Text>
      ),
      responsive: ["sm"],
      title: REVOCATION,
    },
    {
      dataIndex: "published",
      key: "published",
      render: (published: AuthCredential["published"]) => (
        <Typography.Text>{published ? "Published" : "Pending"}</Typography.Text>
      ),
      responsive: ["sm"],
      title: "Published",
    },
    {
      dataIndex: "id",
      key: "id",
      render: (id: AuthCredential["id"], credential: AuthCredential) => (
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
      style={{ width: "100%" }}
      title={
        <Flex align="center" justify="flex-end" style={{ padding: 12 }}>
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
              <Card.Meta title="Auth credentials" />

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
