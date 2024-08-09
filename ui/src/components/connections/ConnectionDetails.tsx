import { Button, Card, Row, Space } from "antd";
import { useCallback, useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";

import { getConnection } from "src/adapters/api/connections";
import IconTrash from "src/assets/icons/trash-01.svg?react";
import { ConnectionDeleteModal } from "src/components/connections/ConnectionDeleteModal";
import { CredentialsTable } from "src/components/connections/CredentialsTable";
import { Detail } from "src/components/shared/Detail";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { useIssuerContext } from "src/contexts/Issuer";
import { AppError, Connection } from "src/domain";
import { AsyncTask, isAsyncTaskDataAvailable } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { DELETE, IDENTIFIER } from "src/utils/constants";
import { formatDate } from "src/utils/forms";

export function ConnectionDetails() {
  const env = useEnvContext();
  const { issuerIdentifier } = useIssuerContext();
  const navigate = useNavigate();

  const [connection, setConnection] = useState<AsyncTask<Connection, AppError>>({
    status: "pending",
  });

  const [showModal, setShowModal] = useState<boolean>(false);

  const { connectionID } = useParams();

  const fetchConnection = useCallback(
    async (signal: AbortSignal) => {
      if (connectionID) {
        setConnection({ status: "loading" });
        const response = await getConnection({
          env,
          id: connectionID,
          issuerIdentifier,
          signal,
        });
        if (response.success) {
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
    [connectionID, env, issuerIdentifier]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchConnection);

    return aborter;
  }, [fetchConnection]);

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
                      <Button
                        danger
                        icon={<IconTrash />}
                        onClick={() => setShowModal(true)}
                        type="text"
                      >
                        {DELETE}
                      </Button>
                    </Row>
                    <Card className="background-grey">
                      <Detail
                        copyable
                        ellipsisPosition={5}
                        label={IDENTIFIER}
                        text={connection.data.userID}
                      />
                      <Detail label="Creation date" text={formatDate(connection.data.createdAt)} />
                    </Card>
                  </Space>
                );
              }
            }
          })()}
        </Card>
        {isAsyncTaskDataAvailable(connection) && (
          <CredentialsTable userID={connection.data.userID} />
        )}
      </Space>
      {connectionID && showModal && (
        <ConnectionDeleteModal
          id={connectionID}
          onClose={() => setShowModal(false)}
          onDelete={() => navigate(-1)}
        />
      )}
    </SiderLayoutContent>
  );
}
