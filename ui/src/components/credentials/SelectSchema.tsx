import { Form, Select, message } from "antd";
import { useCallback, useEffect, useState } from "react";

import { generatePath, useNavigate } from "react-router-dom";
import { Schema, schemasGetAll } from "src/adapters/api/schemas";
import { useEnvContext } from "src/contexts/env";
import { ROUTES } from "src/routes";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { SCHEMA_TYPE } from "src/utils/constants";
import { AsyncTask, isAsyncTaskDataAvailable } from "src/utils/types";

export function SelectSchema({ schemaID }: { schemaID: string | undefined }) {
  const env = useEnvContext();
  const [schemas, setSchemas] = useState<AsyncTask<Schema[], undefined>>({
    status: "pending",
  });

  const navigate = useNavigate();

  const getSchemas = useCallback(
    async (signal: AbortSignal) => {
      setSchemas((oldState) =>
        isAsyncTaskDataAvailable(oldState)
          ? { data: oldState.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await schemasGetAll({
        env,
        params: {},
        signal,
      });

      if (response.isSuccessful) {
        setSchemas({ data: response.data.schemas, status: "successful" });
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
    const { aborter } = makeRequestAbortable(getSchemas);

    return aborter;
  }, [getSchemas]);

  return (
    <Form layout="vertical">
      <Form.Item label="Select schema type" required>
        <Select
          className="full-width"
          loading={!isAsyncTaskDataAvailable(schemas)}
          onChange={(id: string) =>
            navigate(generatePath(ROUTES.issueCredential.path, { schemaID: id }))
          }
          placeholder={SCHEMA_TYPE}
          value={schemaID && isAsyncTaskDataAvailable(schemas) ? schemaID : undefined}
        >
          {isAsyncTaskDataAvailable(schemas) &&
            schemas.data.map(({ id, schema }) => (
              <Select.Option key={id} value={id}>
                {schema}
              </Select.Option>
            ))}
        </Select>
      </Form.Item>
    </Form>
  );
}
