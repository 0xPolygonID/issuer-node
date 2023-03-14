import { Button, Drawer, Row, Space, Tooltip, Typography } from "antd";
import { useCallback, useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";

import { Schema, schemasGetSingle } from "src/adapters/api/schemas";
import { ReactComponent as IconInfo } from "src/assets/icons/info-circle.svg";
import { ReactComponent as IconClose } from "src/assets/icons/x.svg";
import { CopyableDetail } from "src/components/schemas/CopyableDetail";
import { ErrorResult } from "src/components/schemas/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { APIError } from "src/utils/adapters";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { DETAILS_MAXIMUM_WIDTH, FORM_LABEL, SCHEMA_ID_SEARCH_PARAM } from "src/utils/constants";
import { formatDate } from "src/utils/forms";
import { AsyncTask } from "src/utils/types";

export function SchemaDetails() {
  const [schema, setSchema] = useState<AsyncTask<Schema, APIError>>({
    status: "pending",
  });
  const [searchParams, setSearchParams] = useSearchParams();
  const schemaID = searchParams.get(SCHEMA_ID_SEARCH_PARAM);

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

  const onClose = () => {
    const params = new URLSearchParams(searchParams);

    params.delete(SCHEMA_ID_SEARCH_PARAM);

    setSearchParams(params);
    setSchema({ status: "pending" });
  };

  useEffect(() => {
    const { aborter } = makeRequestAbortable(getSchema);

    return aborter;
  }, [getSchema]);

  return (
    <Drawer
      closable={false}
      extra={<Button icon={<IconClose />} onClick={onClose} size="small" type="text" />}
      maskClosable
      onClose={onClose}
      open={schemaID !== null}
      title="View claim schema"
    >
      {(() => {
        switch (schema.status) {
          case "failed": {
            return <ErrorResult error={schema.error.message} />;
          }
          case "pending":
          case "loading": {
            return <LoadingResult />;
          }
          case "reloading":
          case "successful": {
            const { attributes, createdAt, id, schema: name, schemaHash, schemaURL } = schema.data;

            return (
              <Space direction="vertical">
                <Row justify="space-between">
                  <Typography.Text type="secondary">{FORM_LABEL.SCHEMA_NAME}</Typography.Text>

                  <Typography.Text
                    ellipsis={{ tooltip: true }}
                    style={{ maxWidth: DETAILS_MAXIMUM_WIDTH }}
                  >
                    {name}
                  </Typography.Text>
                </Row>

                <Row justify="space-between">
                  <Typography.Text type="secondary">{FORM_LABEL.ATTRIBUTES}</Typography.Text>

                  <Space direction="vertical" style={{ maxWidth: DETAILS_MAXIMUM_WIDTH }}>
                    {attributes.map(({ description, name }) => (
                      <Row key={name} style={{ justifyContent: "end" }}>
                        <Space>
                          <Typography.Text
                            ellipsis={{ tooltip: true }}
                            style={{ maxWidth: DETAILS_MAXIMUM_WIDTH - 32 }}
                          >
                            {name}
                          </Typography.Text>

                          <Tooltip placement="leftTop" title={description || "No description."}>
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
                  <Typography.Text type="secondary">{FORM_LABEL.CREATION_DATE}</Typography.Text>

                  <Typography.Text ellipsis style={{ maxWidth: DETAILS_MAXIMUM_WIDTH }}>
                    {formatDate(createdAt)}
                  </Typography.Text>
                </Row>

                <CopyableDetail data={id} label={FORM_LABEL.SCHEMA_ID} />

                <CopyableDetail data={schemaHash} label={FORM_LABEL.SCHEMA_HASH} />

                <CopyableDetail data={schemaURL} label="Schema URL" />
              </Space>
            );
          }
        }
      })()}
    </Drawer>
  );
}
