import { Button, Drawer, Row, Space, Tooltip, Typography } from "antd";
import { useCallback, useEffect, useState } from "react";
import { generatePath, useSearchParams } from "react-router-dom";

import { Claim, ClaimAttribute, claimsGetSingle } from "src/adapters/api/claims";
import { formatAttributeValue } from "src/adapters/parsers/forms";
import { ReactComponent as IconCheckMark } from "src/assets/icons/check.svg";
import { ReactComponent as IconCopy } from "src/assets/icons/copy-01.svg";
import { ReactComponent as IconInfo } from "src/assets/icons/info-circle.svg";
import { ReactComponent as IconClose } from "src/assets/icons/x.svg";
import { CopyableDetail } from "src/components/schemas/CopyableDetail";
import { ErrorResult } from "src/components/schemas/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { ROUTES } from "src/routes";
import { APIError } from "src/utils/adapters";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import {
  CLAIM_ID_SEARCH_PARAM,
  DETAILS_MAXIMUM_WIDTH,
  FORM_LABEL,
  SCHEMA_ID_SEARCH_PARAM,
} from "src/utils/constants";
import { formatDate } from "src/utils/forms";
import { AsyncTask, isAsyncTaskDataAvailable } from "src/utils/types";

export function IssuedClaimDetails() {
  const [claim, setClaim] = useState<AsyncTask<Claim, APIError>>({
    status: "pending",
  });

  const [searchParams, setSearchParams] = useSearchParams();
  const claimID = searchParams.get(CLAIM_ID_SEARCH_PARAM);
  const schemaID = searchParams.get(SCHEMA_ID_SEARCH_PARAM);

  const getAttributeDescription = (attributeKey: string): string => {
    const attribute =
      isAsyncTaskDataAvailable(claim) &&
      claim.data.schemaTemplate.attributes.find((attribute) => attribute.name === attributeKey);

    return (attribute && attribute.description) || "No description.";
  };

  const getAttributeValue = (attribute: ClaimAttribute): string => {
    if (isAsyncTaskDataAvailable(claim)) {
      const value = formatAttributeValue(attribute, claim.data.schemaTemplate.attributes);

      return value.success ? value.data : value.error;
    }

    return "-";
  };

  const getClaim = useCallback(
    async (signal: AbortSignal) => {
      if (claimID && schemaID) {
        setClaim({ status: "loading" });

        const response = await claimsGetSingle({
          claimID,
          schemaID,
          signal,
        });

        if (response.isSuccessful) {
          setClaim({ data: response.data, status: "successful" });
        } else {
          if (!isAbortedError(response.error)) {
            setClaim({ error: response.error, status: "failed" });
          }
        }
      }
    },
    [claimID, schemaID]
  );

  const claimLinkPath =
    isAsyncTaskDataAvailable(claim) &&
    generatePath(ROUTES.claimLink.path, {
      claimID: claim.data.id,
    });

  const onClose = () => {
    const params = new URLSearchParams(searchParams);

    params.delete(CLAIM_ID_SEARCH_PARAM);
    params.delete(SCHEMA_ID_SEARCH_PARAM);

    setSearchParams(params);

    setClaim({ status: "pending" });
  };

  useEffect(() => {
    const { aborter } = makeRequestAbortable(getClaim);

    return aborter;
  }, [getClaim]);

  return (
    <Drawer
      closable={false}
      extra={<Button icon={<IconClose />} onClick={onClose} size="small" type="text" />}
      maskClosable
      onClose={onClose}
      open={claimID !== null && schemaID !== null}
      title="Issued claim"
    >
      {(() => {
        switch (claim.status) {
          case "failed": {
            return <ErrorResult error={claim.error.message} />;
          }
          case "loading":
          case "pending": {
            return <LoadingResult />;
          }
          case "reloading":
          case "successful": {
            const {
              attributeValues,
              claimLinkExpiration,
              createdAt,
              expiresAt,
              issuedClaims,
              limitedClaims,
            } = claim.data;
            const { id, schema: schemaName, schemaHash } = claim.data.schemaTemplate;

            const left =
              issuedClaims !== null && limitedClaims !== null ? limitedClaims - issuedClaims : null;
            const numberOfClaimsLeft =
              limitedClaims !== null && left !== null ? `${left} of ${limitedClaims}` : "-";

            return (
              <Space direction="vertical">
                <Row justify="space-between">
                  <Typography.Text type="secondary">{FORM_LABEL.SCHEMA_NAME}</Typography.Text>

                  <Typography.Text
                    ellipsis={{ tooltip: true }}
                    style={{ maxWidth: DETAILS_MAXIMUM_WIDTH }}
                  >
                    {schemaName}
                  </Typography.Text>
                </Row>

                <Row justify="space-between">
                  <Typography.Text type="secondary">{FORM_LABEL.ATTRIBUTES}</Typography.Text>

                  <Space direction="vertical" style={{ maxWidth: DETAILS_MAXIMUM_WIDTH }}>
                    {attributeValues.map(({ attributeKey, attributeValue }) => (
                      <Row key={attributeKey} style={{ justifyContent: "end" }}>
                        <Space>
                          <Typography.Text
                            ellipsis={{ tooltip: true }}
                            style={{ maxWidth: DETAILS_MAXIMUM_WIDTH - 32 }}
                          >
                            {attributeKey} ({getAttributeValue({ attributeKey, attributeValue })})
                          </Typography.Text>

                          <Tooltip
                            placement="leftTop"
                            title={getAttributeDescription(attributeKey)}
                          >
                            <Row>
                              <IconInfo className="hoverable" style={{ verticalAlign: "middle" }} />
                            </Row>
                          </Tooltip>
                        </Space>
                      </Row>
                    ))}
                  </Space>
                </Row>

                <Row justify="space-between">
                  <Typography.Text type="secondary">{FORM_LABEL.LINK_VALIDITY}</Typography.Text>

                  <Typography.Text ellipsis style={{ maxWidth: DETAILS_MAXIMUM_WIDTH }}>
                    {claimLinkExpiration ? formatDate(claimLinkExpiration, true) : "-"}
                  </Typography.Text>
                </Row>

                <Row justify="space-between">
                  <Typography.Text type="secondary">
                    {FORM_LABEL.CLAIM_AVAILABILITY}
                  </Typography.Text>

                  <Typography.Text ellipsis style={{ maxWidth: DETAILS_MAXIMUM_WIDTH }}>
                    {numberOfClaimsLeft}
                  </Typography.Text>
                </Row>

                <Row justify="space-between">
                  <Typography.Text type="secondary">{FORM_LABEL.CREATION_DATE}</Typography.Text>

                  <Typography.Text ellipsis style={{ maxWidth: DETAILS_MAXIMUM_WIDTH }}>
                    {formatDate(createdAt)}
                  </Typography.Text>
                </Row>

                <Row justify="space-between">
                  <Typography.Text type="secondary">{FORM_LABEL.EXPIRATION_DATE}</Typography.Text>

                  <Typography.Text ellipsis style={{ maxWidth: DETAILS_MAXIMUM_WIDTH }}>
                    {expiresAt ? formatDate(expiresAt) : "-"}
                  </Typography.Text>
                </Row>

                <CopyableDetail data={id} label={FORM_LABEL.SCHEMA_ID} />

                <CopyableDetail data={schemaHash} label={FORM_LABEL.SCHEMA_HASH} />

                {claimLinkPath && (
                  <Row justify="space-between">
                    <Typography.Text type="secondary">{FORM_LABEL.CLAIM_LINK}</Typography.Text>

                    <Typography.Link
                      copyable={{ icon: [<IconCopy key={0} />, <IconCheckMark key={1} />] }}
                      ellipsis
                      href={claimLinkPath}
                      style={{ maxWidth: DETAILS_MAXIMUM_WIDTH }}
                      target="_blank"
                    >
                      {`${window.location.origin}${claimLinkPath}`}
                    </Typography.Link>
                  </Row>
                )}
              </Space>
            );
          }
        }
      })()}
    </Drawer>
  );
}
