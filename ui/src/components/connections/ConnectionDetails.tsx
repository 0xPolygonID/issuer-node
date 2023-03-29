import { Avatar, Button, Card, Row, Space, Tag, Typography, message } from "antd";
import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";

import { APIError } from "src/adapters/api";
import { Connection, getConnection } from "src/adapters/api/connections";
import { ReactComponent as IconCheckMark } from "src/assets/icons/check.svg";
import { ReactComponent as IconCopy } from "src/assets/icons/copy-01.svg";
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

  const onClickToImplement = () => {
    void message.error("To be implemented");
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
      } else {
        setConnection({ error: { message: "No ConnectionID" }, status: "failed" });
      }
    },
    [connectionID, env]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchConnection);

    return aborter;
  }, [fetchConnection]);

  const credentialsList =
    isAsyncTaskDataAvailable(connection) && connection.data.credentials
      ? connection.data.credentials
      : [];
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
                      <Button danger icon={<IconTrash />} onClick={onClickToImplement} type="text">
                        Delete connection
                      </Button>
                    </Row>
                    <Card className="background-grey">
                      <Row justify="space-between">
                        <Typography.Text type="secondary">{IDENTIFIER}</Typography.Text>
                        <Typography.Text
                          copyable={{ icon: [<IconCopy key={0} />, <IconCheckMark key={1} />] }}
                        >
                          {connection.data.userID}
                        </Typography.Text>
                      </Row>
                      <Row justify="space-between">
                        <Typography.Text type="secondary">Creation date</Typography.Text>
                        <Typography.Text>
                          {formatDate(connection.data.createdAt, true)}
                        </Typography.Text>
                      </Row>
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
            onSearch={onClickToImplement}
            query="To be implemented"
            searchPlaceholder="Search credentials, attributes..."
            showDefaultContents={connection.status === "successful" && credentialsList.length === 0}
            table={<></>}
            title={
              <Row align="middle" justify="space-between">
                <Space align="end" size="middle">
                  <Card.Meta title={ISSUE_CREDENTIAL} />

                  <Tag color="blue">{credentialsList.length}</Tag>
                </Space>
                <Button
                  icon={<IconCreditCardPlus />}
                  onClick={() => void message.error("To be implemented")}
                  type="primary"
                >
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
