import { Button, Card, Space, TagProps, Typography } from "antd";
import { useCallback, useEffect, useState } from "react";
import { generatePath, useNavigate, useParams } from "react-router-dom";
import { z } from "zod";

import { APIError } from "src/adapters/api";
import { getLink } from "src/adapters/api/credentials";
import { getJsonSchemaFromUrl } from "src/adapters/jsonSchemas";
import { getAttributeValueParser } from "src/adapters/parsers/jsonSchemas";
import { ReactComponent as IconTrash } from "src/assets/icons/trash-01.svg";
import { LinkDeleteModal } from "src/components/credentials/LinkDeleteModal";
import { ObjectAttributeValueTree } from "src/components/credentials/ObjectAttributeValueTree";
import { Detail } from "src/components/shared/Detail";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { Link, ObjectAttribute, ObjectAttributeValue } from "src/domain";
import { ROUTES } from "src/routes";
import {
  AsyncTask,
  hasAsyncTaskFailed,
  isAsyncTaskDataAvailable,
  isAsyncTaskStarting,
} from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { CREDENTIALS_TABS } from "src/utils/constants";
import { processError, processZodError } from "src/utils/error";
import { formatDate } from "src/utils/forms";

export function LinkDetails() {
  const navigate = useNavigate();
  const { linkID } = useParams();

  const env = useEnvContext();

  const [credentialSubjectValue, setCredentialSubjectValue] = useState<
    AsyncTask<ObjectAttributeValue, string | z.ZodError>
  >({
    status: "pending",
  });
  const [link, setLink] = useState<AsyncTask<Link, APIError>>({
    status: "pending",
  });

  const [showModal, setShowModal] = useState<boolean>(false);

  const fetchJsonSchemaFromUrl = useCallback(({ link }: { link: Link }): void => {
    setCredentialSubjectValue({ status: "loading" });

    getJsonSchemaFromUrl({ url: link.schemaUrl })
      .then(([jsonSchema]) => {
        const credentialSubjectSchema =
          (jsonSchema.type === "object" &&
            jsonSchema.schema.properties
              ?.filter((child): child is ObjectAttribute => child.type === "object")
              .find((child) => child.name === "credentialSubject")) ||
          null;

        const credentialSubjectSchemaWithoutId: ObjectAttribute | null =
          credentialSubjectSchema && {
            ...credentialSubjectSchema,
            schema: {
              ...credentialSubjectSchema.schema,
              properties: credentialSubjectSchema.schema.properties?.filter(
                (attribute) => attribute.name !== "id"
              ),
            },
          };

        if (credentialSubjectSchemaWithoutId) {
          const parsedCredentialSubject = getAttributeValueParser(
            credentialSubjectSchemaWithoutId
          ).safeParse(link.credentialSubject);

          if (parsedCredentialSubject.success) {
            if (parsedCredentialSubject.data.type === "object") {
              setCredentialSubjectValue({
                data: parsedCredentialSubject.data,
                status: "successful",
              });
            } else {
              setCredentialSubjectValue({
                error: `The type "${parsedCredentialSubject.data.type}" is not a valid type for the attribute "credentialSubject".`,
                status: "failed",
              });
            }
          } else {
            setCredentialSubjectValue({
              error: parsedCredentialSubject.error,
              status: "failed",
            });
          }
        } else {
          setCredentialSubjectValue({
            error: `Could not find the attribute "credentialSubject" in the object's schema.`,
            status: "failed",
          });
        }
      })
      .catch((error) => {
        setCredentialSubjectValue({
          error: processError(error),
          status: "failed",
        });
      });
  }, []);

  const fetchLink = useCallback(
    async (signal: AbortSignal) => {
      if (linkID) {
        setLink({ status: "loading" });

        const response = await getLink({
          env,
          linkID,
          signal,
        });

        if (response.isSuccessful) {
          setLink({ data: response.data, status: "successful" });
          fetchJsonSchemaFromUrl({ link: response.data });
        } else {
          if (!isAbortedError(response.error)) {
            setLink({ error: response.error, status: "failed" });
          }
        }
      }
    },
    [env, fetchJsonSchemaFromUrl, linkID]
  );

  useEffect(() => {
    if (linkID) {
      const { aborter } = makeRequestAbortable(fetchLink);
      return aborter;
    }
    return;
  }, [fetchLink, linkID]);

  const loading = isAsyncTaskStarting(link) || isAsyncTaskStarting(credentialSubjectValue);

  const credentialSubjectValueErrorToString = (error: string | z.ZodError) =>
    error instanceof z.ZodError
      ? [
          "An error occurred while parsing the value of the credentialSubject:",
          ...processZodError(error).map((e) => `"${e}"`),
        ].join("\n")
      : `An error occurred while processing the value of the credentialSubject:\n"${error}"`;

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
          const { createdAt, expiration, proofTypes, schemaHash, schemaType, status } = link.data;

          const linkURL = `${window.location.origin}${generatePath(ROUTES.credentialLinkQR.path, {
            linkID,
          })}`;

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
                  Delete link
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
                      text={expiration ? formatDate(expiration) : "-"}
                    />

                    <Detail copyable label="Schema hash" text={schemaHash} />

                    <Detail copyable label="Link" text={linkURL} />
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
