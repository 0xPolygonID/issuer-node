import { App, Row, Typography } from "antd";

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
  const { message } = App.useApp();

  const handleDownload = () => {
    downloadJsonFromUrl({
      env,
      fileName,
      url,
    })
      .then(() => {
        void message.success("Schema downloaded successfully.");
      })
      .catch(() => {
        void message.error("An error occurred while downloading the schema. Please try again.");
      });
  };

  return (
    <Row justify="space-between">
      <Typography.Text type="secondary">Download</Typography.Text>

      <Typography.Link onClick={handleDownload}>JSON Schema</Typography.Link>
    </Row>
  );
}
