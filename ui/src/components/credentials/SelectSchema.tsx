import { Form, Select, message } from "antd";
import { useCallback, useEffect, useState } from "react";

import { generatePath, useNavigate } from "react-router-dom";
import { getSchemas } from "src/adapters/api/schemas";
import { useEnvContext } from "src/contexts/Env";
import { Schema } from "src/domain";
import { ROUTES } from "src/routes";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { SCHEMA_TYPE } from "src/utils/constants";

export function SelectSchema({
  onSelect,
  schemaID,
}: {
  onSelect: (schema: Schema) => void;
  schemaID: string | undefined;
}) {
  const env = useEnvContext();
  const [schemas, setSchemas] = useState<AsyncTask<Schema[], undefined>>({
    status: "pending",
  });

  const navigate = useNavigate();

  const fetchSchemas = useCallback(
    async (signal: AbortSignal) => {
      setSchemas((previousState) =>
        isAsyncTaskDataAvailable(previousState)
          ? { data: previousState.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await getSchemas({
        env,
        params: {},
        signal,
      });

      if (response.isSuccessful) {
        setSchemas({ data: response.data.successful, status: "successful" });
      } else {
        if (!isAbortedError(response.error)) {
          setSchemas({ error: undefined, status: "failed" });
          void message.error(response.error.message);
        }
      }
    },
    [env]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchSchemas);

    return aborter;
  }, [fetchSchemas]);

  useEffect(() => {
    if (schemaID) {
      const schema =
        isAsyncTaskDataAvailable(schemas) && schemas.data.find((schema) => schema.id === schemaID);
      if (schema) {
        onSelect(schema);
      }
    }
  }, [onSelect, schemaID, schemas]);

  return (
    <Form layout="vertical">
      <Form.Item label="Select schema type" required>
        <Select
          className="full-width"
          loading={isAsyncTaskStarting(schemas)}
          onChange={(id: string) => {
            const schema =
              isAsyncTaskDataAvailable(schemas) && schemas.data.find((schema) => schema.id === id);
            if (schema) {
              navigate(generatePath(ROUTES.issueCredential.path, { schemaID: id }), {
                replace: true,
              });
              onSelect(schema);
            }
          }}
          placeholder={SCHEMA_TYPE}
          value={schemaID && isAsyncTaskDataAvailable(schemas) ? schemaID : undefined}
        >
          {isAsyncTaskDataAvailable(schemas) &&
            schemas.data.map(({ id, type }) => (
              <Select.Option key={id} value={id}>
                {type}
              </Select.Option>
            ))}
        </Select>
      </Form.Item>
    </Form>
  );
}
