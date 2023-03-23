import { Form, Select, message } from "antd";
import { useCallback, useEffect, useState } from "react";

import { generatePath, useNavigate } from "react-router-dom";
import { Schema, getSchemas } from "src/adapters/api/schemas";
import { ROUTES } from "src/routes";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { SCHEMA_TYPE } from "src/utils/constants";
import { AsyncTask, isAsyncTaskDataAvailable } from "src/utils/types";

export function SelectSchema({ schemaID }: { schemaID: string | undefined }) {
  const [schemas, setSchemas] = useState<AsyncTask<Schema[], undefined>>({
    status: "pending",
  });

  const navigate = useNavigate();

  const fetchSchemas = useCallback(async (signal: AbortSignal) => {
    setSchemas((oldState) =>
      isAsyncTaskDataAvailable(oldState)
        ? { data: oldState.data, status: "reloading" }
        : { status: "loading" }
    );

    const response = await getSchemas({
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
  }, []);

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
