import { Button, Result } from "antd";
import { Link } from "react-router-dom";

import { ROOT_PATH } from "src/utils/constants";

export function NotFound() {
  return (
    <Result
      extra={
        <Link to={ROOT_PATH}>
          <Button type="primary">Go to home</Button>
        </Link>
      }
      status="warning"
      subTitle="This page does not exist."
      title="404"
    />
  );
}
