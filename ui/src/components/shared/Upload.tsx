import { Button, Upload, UploadProps } from "antd";
import { ReactComponent as IconUpload } from "src/assets/icons/upload-01.svg";

export function UploadDoc() {
  const props: UploadProps = {
    action: "https://run.mocky.io/v3/435e224c-44fb-4773-9faf-380c5e6a2188",
    headers: {
      authorization: "authorization-text",
    },
    maxCount: 1,
    name: "file",
    onChange(info) {
      if (info.file.status !== "uploading") {
        console.log(info.file, info.fileList);
      }
      if (info.file.status === "done") {
        // message.success(`${info.file.name} file uploaded successfully`);
      } else if (info.file.status === "error") {
        //message.error(`${info.file.name} file upload failed.`);
      }
    },
  };
  return (
    <>
      <Upload {...props}>
        <Button icon={<IconUpload />} style={{ height: 20, width: 40 }} type="primary" />
      </Upload>
    </>
  );
}
