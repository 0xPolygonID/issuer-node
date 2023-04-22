import { Button, Card, Space, Typography } from "antd";
import { useCallback, useEffect, useState } from "react";
import { generatePath, useNavigate, useParams } from "react-router-dom";
import { z } from "zod";

import { LinkDeleteModal } from "./LinkDeleteModal";
import { APIError } from "src/adapters/api";
import { getLink } from "src/adapters/api/credentials";
import { getJsonSchemaFromUrl } from "src/adapters/jsonSchemas";
import { getAttributeValueParser } from "src/adapters/parsers/jsonSchemas";
import { ReactComponent as IconTrash } from "src/assets/icons/trash-01.svg";
import { ObjectAttributeValuesTree } from "src/components/credentials/ObjectAttributeValuesTree";
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
import { processZodError } from "src/utils/error";
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

    void getJsonSchemaFromUrl({ url: link.schemaUrl }).then(([jsonSchema]) => {
      const credentialSubject =
        (jsonSchema.type === "object" &&
          jsonSchema.schema.properties
            ?.filter((child): child is ObjectAttribute => child.type === "object")
            .find((child) => child.name === "credentialSubject")) ||
        null;

      if (credentialSubject) {
        const parsedCredentialSubject = getAttributeValueParser(credentialSubject).safeParse(
          link.credentialSubject
        );

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

  const jsonSchemaErrorToString = (error: string | z.ZodError) =>
    error instanceof z.ZodError
      ? [
          "An error occurred while parsing the schema from the URL:",
          ...processZodError(error).map((e) => `"${e}"`),
          "Please provide a valid schema.",
        ].join("\n")
      : `An error occurred while downloading the schema from the URL:\n"${error}"\nPlease try again.`;

  return (
    <SiderLayoutContent
      description="Control credential link accessibility, add notes and change settings."
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
              <ErrorResult error={jsonSchemaErrorToString(credentialSubjectValue.error)} />
            </Card>
          );
        } else if (loading) {
          return (
            <Card className="centered">
              <LoadingResult />
            </Card>
          );
        } else {
          const { expiration, proofTypes, schemaType, status } = link.data;

          const linkURL = `${window.location.origin}${generatePath(ROUTES.credentialLink.path, {
            linkID,
          })}`;

          const [flavor, text] = (() => {
            switch (status) {
              case "active": {
                return [{ color: "success", type: "tag" } as const, "Active"];
              }
              case "inactive": {
                return [{ type: "tag" } as const, "Inactive"];
              }
              case "exceeded": {
                return [{ color: "error", type: "tag" } as const, "Exceeded"];
              }
            }
          })();

          return (
            <Card
              className="centered"
              extra={
                <Button danger icon={<IconTrash />} onClick={() => setShowModal(true)} type="text">
                  Delete Link
                </Button>
              }
              title={schemaType}
            >
              <Space direction="vertical" size="large">
                <Card className="background-grey">
                  <Space direction="vertical">
                    <Typography.Text type="secondary">CREDENTIAL LINK DETAILS</Typography.Text>

                    <Detail flavor={flavor} label="Link status" text={text} />

                    <Detail label="Proof type" text={proofTypes.join(", ")} />

                    <Detail label="Creation date" text="-" />

                    <Detail
                      label="Credential expiration date"
                      text={expiration ? formatDate(expiration, "date-time") : "-"}
                    />

                    <Detail copyable label="Schema hash" text="-" />

                    <Detail copyable label="Link" text={linkURL} />
                  </Space>
                </Card>
                <Card className="background-grey">
                  <Space direction="vertical" size="middle">
                    <Typography.Text type="secondary">ATTRIBUTES</Typography.Text>

                    <ObjectAttributeValuesTree
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
