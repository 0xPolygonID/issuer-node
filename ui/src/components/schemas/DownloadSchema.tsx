import { Row, Typography, message } from "antd";

import { downloadJsonFromUrl } from "src/adapters/json";
import { Env } from "src/domain";

export function DownloadSchema({
  env,
  fileName,
  url,
}: {
  env: Env;
  fileName: string;
  url: string;
}) {
  const [messageAPI, messageContext] = message.useMessage();

  const handleDownload = () => {
    downloadJsonFromUrl({
      env,
      fileName,
      url,
    })
      .then(() => {
        void messageAPI.success("Schema downloaded successfully.");
      })
      .catch(() => {
        void messageAPI.error("An error occurred while downloading the schema. Please try again.");
      });
  };

  return (
    <>
      {messageContext}

      <Row justify="space-between">
        <Typography.Text type="secondary">Download</Typography.Text>

        <Typography.Link onClick={handleDownload}>JSON Schema</Typography.Link>
      </Row>
    </>
  );
}
