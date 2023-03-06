import { message } from "antd";
import { useCallback, useEffect, useState } from "react";
import { generatePath, useParams } from "react-router-dom";

import { Claim, claimIssue } from "src/adapters/api/claims";
import { Schema, schemasGetSingle } from "src/adapters/api/schemas";
import { issueClaimFormData } from "src/adapters/parsers/forms";
import { serializeClaimForm } from "src/adapters/parsers/serializers";
import { ErrorResult } from "src/components/schemas/ErrorResult";
import { AttributeValues, SetAttributes } from "src/components/schemas/SetAttributes";
import { IssuanceMethod, SetIssuanceMethod } from "src/components/schemas/SetIssuanceMethod";
import { Summary } from "src/components/schemas/Summary";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { ROUTES } from "src/routes";
import { APIError, processZodError } from "src/utils/adapters";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { SCHEMAS_TABS } from "src/utils/constants";
import { AsyncTask, isAsyncTaskDataAvailable } from "src/utils/types";

type IssuanceStep = "attributes" | "issuanceMethod" | "summary";

interface FormData {
  attributes: AttributeValues;
  issuanceMethod: IssuanceMethod;
}

export function Issuance() {
  const [claim, setClaim] = useState<AsyncTask<Claim, undefined>>({
    status: "pending",
  });
  const [formData, setFormData] = useState<FormData>({
    attributes: {},
    issuanceMethod: {
      type: "claimLink",
    },
  });
  const [schema, setSchema] = useState<AsyncTask<Schema, APIError>>({
    status: "pending",
  });
  const [step, setStep] = useState<IssuanceStep>("attributes");

  const { schemaID } = useParams();

  const getSchema = useCallback(
    async (signal: AbortSignal) => {
      if (schemaID) {
        setSchema({ status: "loading" });

        const response = await schemasGetSingle({
          schemaID,
          signal,
        });
        if (response.isSuccessful) {
          setSchema({ data: response.data, status: "successful" });
        } else {
          if (!isAbortedError(response.error)) {
            setSchema({ error: response.error, status: "failed" });
          }
        }
      }
    },
    [schemaID]
  );

  const issueClaim = (formData: FormData, schema: Schema) => {
    if (schemaID) {
      const parsedForm = issueClaimFormData(schema.attributes).safeParse(formData);

      if (parsedForm.success) {
          void claimIssue({
            payload: serializeClaimForm(parsedForm.data),
            schemaID,
          }).then((response) => {
            if (response.isSuccessful) {
              setClaim({ data: response.data, status: "successful" });
              void message.success("Claim link created");
              setStep("summary");
            } else {
              setClaim({ error: undefined, status: "failed" });
              void message.error(response.error.message);
            }
          });
      } else {
        processZodError(parsedForm.error).forEach((msg) => void message.error(msg));
      }
    }
  };

  useEffect(() => {
    const { aborter } = makeRequestAbortable(getSchema);
    return aborter;
  }, [getSchema]);

  return (
    <SiderLayoutContent
      backButtonLink={generatePath(ROUTES.schemas.path, {
        tabID: SCHEMAS_TABS[0].tabID,
      })}
      description="A claim is issued with the attribute's value and will generate a QR code for scanning with the Polygon ID wallet."
      showDivider
      title="Issue claims"
    >
      {(() => {
        switch (schema.status) {
          case "failed": {
            return <ErrorResult error={schema.error.message} />;
          }
          case "loading":
          case "pending": {
            return <LoadingResult />;
          }
          case "reloading":
          case "successful": {
            switch (step) {
              case "attributes": {
                return (
                  <SetAttributes
                    initialValues={formData.attributes}
                    onSubmit={(values) => {
                      const updatedValues = values.expirationDate
                        ? { ...values, expirationDate: values.expirationDate.endOf("day") }
                        : values;
                      const newFormData: FormData =
                        formData.issuanceMethod.type === "claimLink" &&
                        updatedValues.expirationDate?.isBefore(
                          formData.issuanceMethod.claimLinkExpirationDate
                        )
                          ? {
                              ...formData,
                              attributes: updatedValues,
                              issuanceMethod: {
                                ...formData.issuanceMethod,
                                claimLinkExpirationDate: undefined,
                                claimLinkExpirationTime: undefined,
                              },
                            }
                          : { ...formData, attributes: updatedValues };

                      setFormData(newFormData);
                      setStep("issuanceMethod");
                    }}
                    schema={schema.data}
                  />
                );
              }

              case "issuanceMethod": {
                return (
                  <SetIssuanceMethod
                    claimExpirationDate={formData.attributes.expirationDate}
                    initialValues={formData.issuanceMethod}
                    isClaimLoading={claim.status === "loading"}
                    onStepBack={(values) => {
                      const newFormData = { ...formData, issuanceMethod: values };

                      setFormData(newFormData);
                      setStep("attributes");
                    }}
                    onSubmit={(values) => {
                      const newFormData = { ...formData, issuanceMethod: values };

                      setFormData(newFormData);
                      issueClaim(newFormData, schema.data);
                    }}
                  />
                );
              }

              case "summary": {
                return (
                  isAsyncTaskDataAvailable(claim) && (
                    <Summary claim={claim.data} schema={schema.data} />
                  )
                );
              }
            }
          }
        }
      })()}
    </SiderLayoutContent>
  );
}
