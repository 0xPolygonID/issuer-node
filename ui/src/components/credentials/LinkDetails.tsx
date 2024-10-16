import { Button, Card, Space, TagProps, Typography } from "antd";
import { useCallback, useEffect, useState } from "react";
import { generatePath, useNavigate, useParams } from "react-router-dom";

import { getLink } from "src/adapters/api/credentials";
import { getJsonSchemaFromUrl } from "src/adapters/jsonSchemas";
import { getAttributeValueParser } from "src/adapters/parsers/jsonSchemas";
import IconTrash from "src/assets/icons/trash-01.svg?react";
import { LinkDeleteModal } from "src/components/credentials/LinkDeleteModal";
import { ObjectAttributeValueTree } from "src/components/credentials/ObjectAttributeValueTree";
import { Detail } from "src/components/shared/Detail";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { AppError, Link, ObjectAttributeValue } from "src/domain";
import { ROUTES } from "src/routes";
import {
  AsyncTask,
  hasAsyncTaskFailed,
  isAsyncTaskDataAvailable,
  isAsyncTaskStarting,
} from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { CREDENTIALS_TABS, DELETE } from "src/utils/constants";
import { buildAppError, credentialSubjectValueErrorToString } from "src/utils/error";
import { formatDate } from "src/utils/forms";
import { extractCredentialSubjectAttributeWithoutId } from "src/utils/jsonSchemas";

export function LinkDetails() {
  const navigate = useNavigate();
  const { linkID } = useParams();

  const env = useEnvContext();
  const { identifier } = useIdentityContext();

  const [credentialSubjectValue, setCredentialSubjectValue] = useState<
    AsyncTask<ObjectAttributeValue, AppError>
  >({
    status: "pending",
  });
  const [link, setLink] = useState<AsyncTask<Link, AppError>>({
    status: "pending",
  });

  const [showModal, setShowModal] = useState<boolean>(false);

  const fetchJsonSchemaFromUrl = useCallback(
    ({ link }: { link: Link }): void => {
      setCredentialSubjectValue({ status: "loading" });

      void getJsonSchemaFromUrl({ env, url: link.schemaUrl }).then((response) => {
        if (response.success) {
          const [jsonSchema] = response.data;

          const credentialSubjectAttributeWithoutId =
            extractCredentialSubjectAttributeWithoutId(jsonSchema);

          if (credentialSubjectAttributeWithoutId) {
            const parsedCredentialSubject = getAttributeValueParser(
              credentialSubjectAttributeWithoutId
            ).safeParse(link.credentialSubject);

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

  const fetchLink = useCallback(
    async (signal: AbortSignal) => {
      if (linkID) {
        setLink({ status: "loading" });

        const response = await getLink({
          env,
          identifier,
          linkID,
          signal,
        });

        if (response.success) {
          setLink({ data: response.data, status: "successful" });
          fetchJsonSchemaFromUrl({ link: response.data });
        } else {
          if (!isAbortedError(response.error)) {
            setLink({ error: response.error, status: "failed" });
          }
        }
      }
    },
    [env, fetchJsonSchemaFromUrl, linkID, identifier]
  );

  useEffect(() => {
    if (linkID) {
      const { aborter } = makeRequestAbortable(fetchLink);
      return aborter;
    }
    return;
  }, [fetchLink, linkID]);

  const loading = isAsyncTaskStarting(link) || isAsyncTaskStarting(credentialSubjectValue);

  return (
    <SiderLayoutContent
      description="View credential link details, attribute values and delete links."
      showBackButton
      showDivider
      title="Credential link details"
    >
      {(() => {
        if (hasAsyncTaskFailed(link)) {
          return (
            <Card className="centered">
              <ErrorResult
                error={[
                  "An error occurred while downloading or parsing the link from the API:",
                  link.error.message,
                ].join("\n")}
              />
            </Card>
          );
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
            createdAt,
            credentialExpiration,
            deepLink,
            proofTypes,
            schemaHash,
            schemaType,
            status,
            universalLink,
          } = link.data;

          const [tag, text]: [TagProps, string] = (() => {
            switch (status) {
              case "active": {
                return [{ color: "success" }, "Active"];
              }
              case "inactive": {
                return [{}, "Inactive"];
              }
              case "exceeded": {
                return [{ color: "error" }, "Exceeded"];
              }
            }
          })();

          return (
            <Card
              className="centered"
              extra={
                <Button danger icon={<IconTrash />} onClick={() => setShowModal(true)} type="text">
                  {DELETE}
                </Button>
              }
              title={schemaType}
            >
              <Space direction="vertical" size="large">
                <Card className="background-grey">
                  <Space direction="vertical">
                    <Typography.Text type="secondary">CREDENTIAL LINK DETAILS</Typography.Text>

                    <Detail label="Link status" tag={tag} text={text} />

                    <Detail label="Proof type" text={proofTypes.join(", ")} />

                    <Detail label="Creation date" text={formatDate(createdAt)} />

                    <Detail
                      label="Credential expiration date"
                      text={credentialExpiration ? formatDate(credentialExpiration, "date") : "-"}
                    />

                    <Detail copyable label="Schema hash" text={schemaHash} />

                    <Detail
                      copyable
                      downloadLink
                      href={universalLink}
                      label="Universal link"
                      text={universalLink}
                    />

                    <Detail
                      copyable
                      downloadLink
                      href={deepLink}
                      label="Deep link"
                      text={deepLink}
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
      {isAsyncTaskDataAvailable(link) && showModal && (
        <LinkDeleteModal
          id={link.data.id}
          onClose={() => setShowModal(false)}
          onDelete={() =>
            navigate(
              generatePath(ROUTES.credentials.path, {
                tabID: CREDENTIALS_TABS[1].tabID,
              })
            )
          }
        />
      )}
    </SiderLayoutContent>
  );
}
