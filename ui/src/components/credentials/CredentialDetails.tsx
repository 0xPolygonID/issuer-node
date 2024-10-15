import { Button, Card, Col, Grid, Row, Space, Typography } from "antd";
import { useCallback, useEffect, useState } from "react";
import { generatePath, useNavigate, useParams } from "react-router-dom";

import { getCredential, getIssuedMessages } from "src/adapters/api/credentials";
import { getJsonSchemaFromUrl } from "src/adapters/jsonSchemas";
import { getAttributeValueParser } from "src/adapters/parsers/jsonSchemas";
import IconTrash from "src/assets/icons/trash-01.svg?react";
import IconClose from "src/assets/icons/x.svg?react";
import { ObjectAttributeValueTree } from "src/components/credentials/ObjectAttributeValueTree";
import { CredentialDeleteModal } from "src/components/shared/CredentialDeleteModal";
import { CredentialRevokeModal } from "src/components/shared/CredentialRevokeModal";
import { Detail } from "src/components/shared/Detail";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { AppError, Credential, IssuedMessage, ObjectAttributeValue } from "src/domain";
import { ROUTES } from "src/routes";
import {
  AsyncTask,
  hasAsyncTaskFailed,
  isAsyncTaskDataAvailable,
  isAsyncTaskStarting,
} from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { CREDENTIALS_TABS, DELETE, REVOKE } from "src/utils/constants";
import { buildAppError, credentialSubjectValueErrorToString } from "src/utils/error";
import { formatDate } from "src/utils/forms";
import { extractCredentialSubjectAttribute } from "src/utils/jsonSchemas";

export function CredentialDetails() {
  const navigate = useNavigate();
  const { credentialID } = useParams();

  const { sm } = Grid.useBreakpoint();

  const env = useEnvContext();
  const { identifier } = useIdentityContext();

  const [credentialSubjectValue, setCredentialSubjectValue] = useState<
    AsyncTask<ObjectAttributeValue, AppError>
  >({
    status: "pending",
  });
  const [credential, setCredential] = useState<AsyncTask<Credential, AppError>>({
    status: "pending",
  });
  const [issuedMessages, setIssuedMessages] = useState<
    AsyncTask<[IssuedMessage, IssuedMessage], AppError>
  >({
    status: "pending",
  });
  const [showDeleteModal, setShowDeleteModal] = useState<boolean>(false);
  const [showRevokeModal, setShowRevokeModal] = useState<boolean>(false);

  const fetchJsonSchemaFromUrl = useCallback(
    ({ credential }: { credential: Credential }): void => {
      setCredentialSubjectValue({ status: "loading" });

      void getJsonSchemaFromUrl({ env, url: credential.schemaUrl }).then((response) => {
        if (response.success) {
          const [jsonSchema] = response.data;
          const credentialSubjectAttribute = extractCredentialSubjectAttribute(jsonSchema);
          if (credentialSubjectAttribute) {
            const parsedCredentialSubject = getAttributeValueParser(
              credentialSubjectAttribute
            ).safeParse(credential.credentialSubject);

            if (parsedCredentialSubject.success) {
              if (parsedCredentialSubject.data.type === "object") {
                setCredentialSubjectValue({
                  data: parsedCredentialSubject.data,
                  status: "successful",
                });
              } else {
                setCredentialSubjectValue({
                  error: buildAppError(
                    `The type "${parsedCredentialSubject.data.type}" is not a valid type for the attribute "credentialSubject".`
                  ),
                  status: "failed",
                });
              }
            } else {
              setCredentialSubjectValue({
                error: buildAppError(parsedCredentialSubject.error),
                status: "failed",
              });
            }
          } else {
            setCredentialSubjectValue({
              error: buildAppError(
                `Could not find the attribute "credentialSubject" in the object's schema.`
              ),
              status: "failed",
            });
          }
        } else {
          setCredentialSubjectValue({
            error: response.error,
            status: "failed",
          });
        }
      });
    },
    [env]
  );

  const fetchCredentialIssuedMessages = useCallback(
    (signal?: AbortSignal) => {
      if (credentialID) {
        setIssuedMessages({ status: "loading" });

        void getIssuedMessages({ credentialID, env, identifier, signal }).then((response) => {
          if (response.success) {
            setIssuedMessages({ data: response.data, status: "successful" });
          } else {
            if (!isAbortedError(response.error)) {
              setIssuedMessages({ error: response.error, status: "failed" });
            }
          }
        });
      }
    },
    [credentialID, env, identifier]
  );

  const fetchCredential = useCallback(
    async (signal?: AbortSignal) => {
      if (credentialID) {
        setCredential({ status: "loading" });

        const response = await getCredential({
          credentialID,
          env,
          identifier,
          signal,
        });

        if (response.success) {
          setCredential({ data: response.data, status: "successful" });
          fetchJsonSchemaFromUrl({ credential: response.data });
          fetchCredentialIssuedMessages(signal);
        } else {
          if (!isAbortedError(response.error)) {
            setCredential({ error: response.error, status: "failed" });
          }
        }
      }
    },
    [env, fetchJsonSchemaFromUrl, fetchCredentialIssuedMessages, credentialID, identifier]
  );

  useEffect(() => {
    if (credentialID) {
      const { aborter } = makeRequestAbortable(fetchCredential);
      return aborter;
    }
    return;
  }, [fetchCredential, credentialID]);

  const loading =
    isAsyncTaskStarting(credential) ||
    isAsyncTaskStarting(credentialSubjectValue) ||
    isAsyncTaskStarting(issuedMessages);

  return (
    <SiderLayoutContent
      description="View credential details, attribute values and revoke credentials."
      showBackButton
      showDivider
      title="Credential details"
    >
      {(() => {
        if (hasAsyncTaskFailed(credential)) {
          return (
            <Card className="centered">
              <ErrorResult
                error={[
                  "An error occurred while downloading or parsing the credential from the API:",
                  credential.error.message,
                ].join("\n")}
              />
            </Card>
          );
        } else if (hasAsyncTaskFailed(issuedMessages)) {
          return <ErrorResult error={issuedMessages.error.message} />;
        } else if (hasAsyncTaskFailed(credentialSubjectValue)) {
          return (
            <Card className="centered">
              <ErrorResult
                error={credentialSubjectValueErrorToString(credentialSubjectValue.error)}
              />
            </Card>
          );
        } else if (loading) {
          return (
            <Card className="centered">
              <LoadingResult />
            </Card>
          );
        } else {
          const {
            expirationDate,
            issuanceDate,
            proofTypes,
            refreshService,
            revoked,
            schemaHash,
            schemaType,
            userID,
          } = credential.data;

          const [universalLinkResponse, deepLinkResponse] = issuedMessages.data;

          return (
            <Card
              className="centered"
              extra={
                <Row gutter={[0, 8]} justify="end">
                  <Col>
                    <Button
                      danger
                      disabled={revoked}
                      icon={<IconClose />}
                      onClick={() => setShowRevokeModal(true)}
                      type="text"
                    >
                      {sm && REVOKE}
                    </Button>
                  </Col>

                  <Col>
                    <Button
                      danger
                      icon={<IconTrash />}
                      onClick={() => setShowDeleteModal(true)}
                      type="text"
                    >
                      {sm && DELETE}
                    </Button>
                  </Col>
                </Row>
              }
              title={schemaType}
            >
              <Space direction="vertical" size="large">
                <Card className="background-grey">
                  <Space direction="vertical">
                    <Typography.Text type="secondary">CREDENTIAL DETAILS</Typography.Text>

                    <Detail label="Proof type" text={proofTypes.join(", ")} />

                    <Detail label="Creation date" text={formatDate(issuanceDate)} />

                    <Detail
                      label="Credential expiration date"
                      text={expirationDate ? formatDate(expirationDate) : "-"}
                    />

                    <Detail
                      label="Refresh Service"
                      text={refreshService ? refreshService.id : "-"}
                    />

                    <Detail
                      copyable
                      ellipsisPosition={5}
                      label="Issued to identifier"
                      text={userID}
                    />

                    <Detail copyable label="Schema hash" text={schemaHash} />

                    <Detail
                      copyable
                      downloadLink
                      href={universalLinkResponse.universalLink}
                      label="Universal link"
                      text={universalLinkResponse.universalLink}
                    />

                    <Detail
                      copyable
                      downloadLink
                      href={deepLinkResponse.universalLink}
                      label="Deep link"
                      text={deepLinkResponse.universalLink}
                    />
                  </Space>
                </Card>

                <Card className="background-grey">
                  <Space direction="vertical" size="middle">
                    <Typography.Text type="secondary">ATTRIBUTES</Typography.Text>

                    <ObjectAttributeValueTree
                      attributeValue={credentialSubjectValue.data}
                      className="background-grey"
                    />
                  </Space>
                </Card>
              </Space>
            </Card>
          );
        }
      })()}
      {isAsyncTaskDataAvailable(credential) && showDeleteModal && (
        <CredentialDeleteModal
          credential={credential.data}
          onClose={() => setShowDeleteModal(false)}
          onDelete={() =>
            navigate(
              generatePath(ROUTES.credentials.path, {
                tabID: CREDENTIALS_TABS[0].tabID,
              })
            )
          }
        />
      )}
      {isAsyncTaskDataAvailable(credential) && showRevokeModal && (
        <CredentialRevokeModal
          credential={credential.data}
          onClose={() => setShowRevokeModal(false)}
          onRevoke={() => void fetchCredential()}
        />
      )}
    </SiderLayoutContent>
  );
}
