import { Result } from "antd";

import { Spinner } from "src/components/shared/Spinner";

export function LoadingResult() {
  return <Result icon={<Spinner />} title="Loading ..." />;
}
