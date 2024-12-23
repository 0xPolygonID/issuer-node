import {
  App,
  Button,
  Card,
  Divider,
  Dropdown,
  Flex,
  Form,
  Input,
  Row,
  Space,
  Typography,
} from "antd";
import { useCallback, useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";

import { z } from "zod";
import { DISPLAY_METHOD_DETAILS, DISPLAY_METHOD_EDIT, VALUE_REQUIRED } from "../../utils/constants";
import {
  UpsertDisplayMethod,
  deleteDisplayMethod,
  getDisplayMethod,
  getDisplayMethodMetadata,
  updateDisplayMethod,
} from "src/adapters/api/display-method";
import { processUrl } from "src/adapters/api/schemas";
import { buildAppError, notifyError } from "src/adapters/parsers";
import IconDots from "src/assets/icons/dots-vertical.svg?react";
import EditIcon from "src/assets/icons/edit-02.svg?react";
import { DisplayMethodCard } from "src/components/display-methods/DisplayMethodCard";
import { DisplayMethodErrorResult } from "src/components/display-methods/DisplayMethodErrorResult";
import { DeleteItem } from "src/components/shared/DeleteItem";
import { Detail } from "src/components/shared/Detail";
import { EditModal } from "src/components/shared/EditModal";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { AppError, DisplayMethod, DisplayMethodMetadata } from "src/domain";
import { ROUTES } from "src/routes";
import {
  AsyncTask,
  hasAsyncTaskFailed,
  isAsyncTaskDataAvailable,
  isAsyncTaskStarting,
} from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";

function Details({ data }: { data: DisplayMethod }) {
  const env = useEnvContext();
  const { name, type, url } = data;

  const processedDisplayMethodUrl = processUrl(url, env);

  return (
    <Space direction="vertical">
      <Typography.Text type="secondary">DISPLAY METHOD DETAILS</Typography.Text>
      <Detail label="Name" text={name} />
      <Detail
        copyable
        href={processedDisplayMethodUrl.success ? processedDisplayMethodUrl.data : url}
        label="URL"
        text={url}
      />
      <Detail label="Type" text={type} />
    </Space>
  );
}

export function DisplayMethodDetails() {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();
  const { displayMethodID } = useParams();
  const { message } = App.useApp();
  const navigate = useNavigate();
  const [form] = Form.useForm<UpsertDisplayMethod>();

  const [isEditModalOpen, setIsEditModalOpen] = useState(false);

  const [displayMethod, setDisplayMethod] = useState<AsyncTask<DisplayMethod, AppError>>({
    status: "pending",
  });

  const [displayMethodMetadata, setDisplayMethodMetadata] = useState<
    AsyncTask<DisplayMethodMetadata, AppError>
  >({
    status: "pending",
  });

  const fetchDisplayMethodMetadata = useCallback(
    (url: string, signal?: AbortSignal) => {
      setDisplayMethodMetadata({ status: "loading" });
      void getDisplayMethodMetadata({
        env,
        signal,
        url,
      }).then((response) => {
        if (response.success) {
          setDisplayMethodMetadata({ data: response.data, status: "successful" });
        } else {
          if (!isAbortedError(response.error)) {
            setDisplayMethodMetadata({ error: response.error, status: "failed" });
          }
        }
      });
    },
    [env]
  );

  const fetchDisplayMethod = useCallback(
    async (signal?: AbortSignal) => {
      setDisplayMethod((previousDisplayMethod) =>
        isAsyncTaskDataAvailable(previousDisplayMethod)
          ? { data: previousDisplayMethod.data, status: "reloading" }
          : { status: "loading" }
      );

      if (!displayMethodID) {
        return;
      }

      const response = await getDisplayMethod({
        displayMethodID,
        env,
        identifier,
        signal,
      });

      if (response.success) {
        setDisplayMethod({
          data: response.data,
          status: "successful",
        });
        fetchDisplayMethodMetadata(response.data.url, signal);
      } else {
        if (!isAbortedError(response.error)) {
          setDisplayMethod({ error: response.error, status: "failed" });
        }
      }
    },
    [env, identifier, displayMethodID, fetchDisplayMethodMetadata]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchDisplayMethod);

    return aborter;
  }, [fetchDisplayMethod]);

  if (!displayMethodID) {
    return <ErrorResult error="No display method provided." />;
  }

  const handleEdit = () => {
    const { name, url } = form.getFieldsValue();
    const parsedUrl = z.string().url().safeParse(url);

    if (parsedUrl.success) {
      void updateDisplayMethod({
        env,
        id: displayMethodID,
        identifier,
        payload: {
          name: name.trim(),
          url: parsedUrl.data,
        },
      }).then((response) => {
        if (response.success) {
          void fetchDisplayMethod();
          setIsEditModalOpen(false);
        } else {
          void notifyError(buildAppError(response.error.message));
        }
      });
    } else {
      void notifyError(buildAppError(`"${url}" is not a valid URL`));
    }
  };

  const handleDeleteDisplayMethod = () => {
    void deleteDisplayMethod({ env, id: displayMethodID, identifier }).then((response) => {
      if (response.success) {
        navigate(ROUTES.displayMethods.path);
        void message.success(response.data.message);
      } else {
        void message.error(response.error.message);
      }
    });
  };

  const editModal = isAsyncTaskDataAvailable(displayMethod) && (
    <EditModal
      onClose={() => setIsEditModalOpen(false)}
      onSubmit={handleEdit}
      open={isEditModalOpen}
      title="Edit display method"
    >
      <Form
        form={form}
        initialValues={{ name: displayMethod.data.name, url: displayMethod.data.url }}
        layout="vertical"
      >
        <Form.Item label="Name" name="name" rules={[{ message: VALUE_REQUIRED, required: true }]}>
          <Input placeholder="Enter name" />
        </Form.Item>

        <Form.Item label="URL" name="url" rules={[{ message: VALUE_REQUIRED, required: true }]}>
          <Input placeholder="Enter URL" />
        </Form.Item>
      </Form>
    </EditModal>
  );

  const cardTitle = isAsyncTaskDataAvailable(displayMethod) && (
    <Flex align="center" gap={8} justify="space-between">
      <Typography.Text style={{ fontWeight: 600 }}>{displayMethod.data.name}</Typography.Text>
      <Flex gap={8}>
        <Button
          icon={<EditIcon />}
          onClick={() => setIsEditModalOpen(true)}
          style={{ flexShrink: 0 }}
          type="text"
        />

        <Dropdown
          menu={{
            items: [
              {
                danger: true,
                key: "delete",
                label: (
                  <DeleteItem
                    onOk={handleDeleteDisplayMethod}
                    title="Are you sure you want to delete this display method?"
                  />
                ),
              },
            ],
          }}
        >
          <Row>
            <IconDots className="icon-secondary" />
          </Row>
        </Dropdown>
      </Flex>
    </Flex>
  );

  return (
    <SiderLayoutContent
      description="View and edit display method details"
      showBackButton
      showDivider
      title={DISPLAY_METHOD_DETAILS}
    >
      {(() => {
        if (hasAsyncTaskFailed(displayMethod)) {
          return (
            <Card className="centered">
              <ErrorResult
                error={[
                  "An error occurred while downloading a display method from the API:",
                  displayMethod.error.message,
                ].join("\n")}
              />
            </Card>
          );
        } else if (hasAsyncTaskFailed(displayMethodMetadata)) {
          return (
            <Card className="centered" title={cardTitle}>
              {isAsyncTaskDataAvailable(displayMethod) && (
                <>
                  <Details data={displayMethod.data} />
                  <Divider />
                  {editModal}
                </>
              )}
              {displayMethodMetadata.error.type === "parse-error" ? (
                <DisplayMethodErrorResult
                  labelRetry={DISPLAY_METHOD_EDIT}
                  message={displayMethodMetadata.error.message}
                  onRetry={() => setIsEditModalOpen(true)}
                />
              ) : (
                <ErrorResult
                  error={[
                    "An error occurred while downloading a display method from the API:",
                    displayMethodMetadata.error.message,
                  ].join("\n")}
                />
              )}
            </Card>
          );
        } else if (
          isAsyncTaskStarting(displayMethod) ||
          isAsyncTaskStarting(displayMethodMetadata)
        ) {
          return (
            <Card className="centered">
              <LoadingResult />
            </Card>
          );
        } else {
          return (
            <Card className="centered" title={cardTitle}>
              <Details data={displayMethod.data} />

              <Divider />

              <Flex justify="center">
                <DisplayMethodCard metadata={displayMethodMetadata.data} />
              </Flex>

              <Divider />

              <Space direction="vertical">
                <Typography.Text type="secondary">DISPLAY METHOD METADATA</Typography.Text>

                <Detail label="Title" text={displayMethodMetadata.data.title} />
                <Detail label="Description" text={displayMethodMetadata.data.description} />
                <Detail label="Issuer name" text={displayMethodMetadata.data.issuerName} />
                <Detail label="Title color" text={displayMethodMetadata.data.titleTextColor} />
                <Detail
                  label="Description color"
                  text={displayMethodMetadata.data.descriptionTextColor}
                />
                <Detail
                  label="Issuer name color"
                  text={displayMethodMetadata.data.issuerTextColor}
                />

                <Detail
                  copyable
                  href={displayMethodMetadata.data.backgroundImageUrl}
                  label="Background image URL"
                  text={displayMethodMetadata.data.backgroundImageUrl}
                />

                <Detail
                  copyable
                  href={displayMethodMetadata.data.logo.uri}
                  label="Logo URL"
                  text={displayMethodMetadata.data.logo.uri}
                />

                <Detail label="Logo alt" text={displayMethodMetadata.data.logo.alt} />
              </Space>
              {editModal}
            </Card>
          );
        }
      })()}
    </SiderLayoutContent>
  );
}
