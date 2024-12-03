import { Button, Divider, Flex, Form, Input } from "antd";
import { useState } from "react";
import { z } from "zod";

import { UpsertDisplayMethod, getDisplayMethodMetadata } from "src/adapters/api/display-method";
import { DisplayMethodErrorResult } from "src/components/display-methods/DisplayMethodErrorResult";
import { useEnvContext } from "src/contexts/Env";
import { AppError, DisplayMethodMetadata } from "src/domain";
import { AsyncTask, hasAsyncTaskFailed, isAsyncTaskStarting } from "src/utils/async";
import { VALUE_REQUIRED } from "src/utils/constants";
import { buildAppError } from "src/utils/error";

export function DisplayMethodForm({
  initialValues,
  onSubmit,
}: {
  initialValues: UpsertDisplayMethod;
  onSubmit: (formValues: UpsertDisplayMethod) => void;
}) {
  const env = useEnvContext();
  const [form] = Form.useForm<UpsertDisplayMethod>();

  const [displayMethodMetadata, setDisplayMethodMetadata] = useState<
    AsyncTask<DisplayMethodMetadata, AppError>
  >({
    status: "pending",
  });

  const fetchMetadata = (formValues: UpsertDisplayMethod) => {
    const { url } = formValues;
    const parsedUrl = z.string().safeParse(url);

    if (parsedUrl.success) {
      void getDisplayMethodMetadata({ env, url: parsedUrl.data }).then((response) => {
        if (response.success) {
          onSubmit(formValues);
        } else {
          setDisplayMethodMetadata({ error: response.error, status: "failed" });
        }
      });
    } else {
      setDisplayMethodMetadata({
        error: buildAppError(`"${url}" is not a valid URL`),
        status: "failed",
      });
    }
  };

  return (
    <>
      {(() => {
        if (isAsyncTaskStarting(displayMethodMetadata)) {
          return (
            <Form
              form={form}
              initialValues={initialValues}
              layout="vertical"
              onFinish={fetchMetadata}
            >
              <Form.Item
                label="Name"
                name="name"
                rules={[{ message: VALUE_REQUIRED, required: true }]}
              >
                <Input placeholder="Enter name" />
              </Form.Item>

              <Form.Item
                label="URL"
                name="url"
                rules={[{ message: VALUE_REQUIRED, required: true }]}
              >
                <Input placeholder="Enter URL" />
              </Form.Item>

              <Divider />

              <Flex justify="flex-end">
                <Button htmlType="submit" type="primary">
                  Submit
                </Button>
              </Flex>
            </Form>
          );
        } else if (hasAsyncTaskFailed(displayMethodMetadata)) {
          return (
            <DisplayMethodErrorResult
              labelRetry="Edit form"
              message={displayMethodMetadata.error.message}
              onRetry={() => setDisplayMethodMetadata({ status: "pending" })}
            />
          );
        } else {
          return;
        }
      })()}
    </>
  );
}
