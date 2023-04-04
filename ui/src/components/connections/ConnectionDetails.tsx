import { Avatar, Button, Card, Row, Space, Tag, Typography } from "antd";
import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";

import { Detail } from "../schemas/Detail";
import { APIError } from "src/adapters/api";
import { Connection, getConnection } from "src/adapters/api/connections";
import { ReactComponent as IconCreditCardPlus } from "src/assets/icons/credit-card-plus.svg";
import { ReactComponent as IconCreditCardRefresh } from "src/assets/icons/credit-card-refresh.svg";
import { ReactComponent as IconTrash } from "src/assets/icons/trash-01.svg";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { TableCard } from "src/components/shared/TableCard";
import { useEnvContext } from "src/contexts/env";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { IDENTIFIER, ISSUE_CREDENTIAL } from "src/utils/constants";
import { formatDate } from "src/utils/forms";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/types";

export function ConnectionDetails() {
  const env = useEnvContext();

  const [connection, setConnection] = useState<AsyncTask<Connection, APIError>>({
    status: "pending",
  });

  const { connectionID } = useParams();

  const obfuscateDID = (did: string) => {
    const didSplit = did.split(":").slice(0, -1).join(":");
    const address = did.split(":").pop();

    return address
      ? `${didSplit}${address.substring(0, 5)}...${address.substring(address.length - 5)}`
      : "-";
  };

  const fetchConnection = useCallback(
    async (signal: AbortSignal) => {
      if (connectionID) {
        setConnection({ status: "loading" });
        const response = await getConnection({
          env,
          id: connectionID,
          signal,
        });
        if (response.isSuccessful) {
          setConnection({
            data: response.data,
            status: "successful",
          });
        } else {
          if (!isAbortedError(response.error)) {
            setConnection({ error: response.error, status: "failed" });
          }
        }
      }
    },
    [connectionID, env]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchConnection);

    return aborter;
  }, [fetchConnection]);

  // TODO PID-481
  const credentialsList = []; /* isAsyncTaskDataAvailable(connection) && connection.data.credentials
      ? connection.data.credentials
      : []; */

  return (
    <SiderLayoutContent
      description="View connection information, credential attribute data. Revoke and delete issued credentials."
      showBackButton
      showDivider
      title="Connection details"
    >
      <Space direction="vertical" size="large">
        <Card>
          {(() => {
            switch (connection.status) {
              case "pending":
              case "loading": {
                return <LoadingResult />;
              }
              case "failed": {
                return <ErrorResult error={connection.error.message} />;
              }
              case "successful":
              case "reloading": {
                return (
                  <Space direction="vertical" size="middle">
                    <Row align="middle" justify="space-between">
                      <Card.Meta title="Connection" />
                      <Button danger icon={<IconTrash />} type="text">
                        Delete connection
                      </Button>
                    </Row>
                    <Card className="background-grey">
                      <Detail
                        copyable
                        data={obfuscateDID(connection.data.userID)}
                        label={IDENTIFIER}
                      />
                      <Detail
                        data={formatDate(connection.data.createdAt, true)}
                        label="Creation date"
                      />
                    </Card>
                  </Space>
                );
              }
            }
          })()}
        </Card>
        {isAsyncTaskDataAvailable(connection) && (
          <TableCard
            defaultContents={
              <>
                <Avatar className="avatar-color-cyan" icon={<IconCreditCardRefresh />} size={48} />

                <Typography.Text strong>No credentials issued</Typography.Text>

                <Typography.Text type="secondary">
                  Credentials for this connection will be listed here.
                </Typography.Text>
              </>
            }
            isLoading={isAsyncTaskStarting(connection)}
            onSearch={() => null}
            query=""
            searchPlaceholder="Search credentials, attributes..."
            showDefaultContents={connection.status === "successful" && credentialsList.length === 0}
            table={<></>}
            title={
              <Row align="middle" justify="space-between">
                <Space align="end" size="middle">
                  <Card.Meta title={ISSUE_CREDENTIAL} />

                  <Tag color="blue">{credentialsList.length}</Tag>
                </Space>
                <Button icon={<IconCreditCardPlus />} type="primary">
                  Issue directly
                </Button>
              </Row>
            }
          />
        )}
      </Space>
    </SiderLayoutContent>
  );
}
