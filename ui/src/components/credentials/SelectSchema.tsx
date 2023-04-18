import { Form, Select, message } from "antd";
import { useCallback, useEffect, useState } from "react";

import { generatePath, useNavigate } from "react-router-dom";
import { getSchemas } from "src/adapters/api/schemas";
import { useEnvContext } from "src/contexts/env";
import { Schema } from "src/domain";
import { ROUTES } from "src/routes";
import { AsyncTask, isAsyncTaskDataAvailable } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { SCHEMA_TYPE } from "src/utils/constants";

export function SelectSchema({ schemaID }: { schemaID: string | undefined }) {
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
    const { aborter } = makeRequestAbortable(fetchSchemas);

    return aborter;
  }, [fetchSchemas]);

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
