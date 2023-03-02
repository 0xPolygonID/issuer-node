import { Avatar, Row, Space, Typography } from "antd";
import { useAuthContext } from "src/hooks/useAuthContext";
import { IMAGE_PLACEHOLDER_PATH } from "src/utils/constants";
import { isAsyncTaskDataAvailable } from "src/utils/types";

export function UserDropdown() {
  const { organization } = useAuthContext();

  const { displayName, logo } = isAsyncTaskDataAvailable(organization)
    ? organization.data
    : {
        displayName: undefined,
        logo: undefined,
      };

  return (
    <Space>
      <Avatar shape="square" size="large" src={logo || IMAGE_PLACEHOLDER_PATH} />

      <Row className="dropdown-info">
        <Typography.Text className="font-small" ellipsis strong>
          {displayName}
        </Typography.Text>
      </Row>
    </Space>
  );
}
