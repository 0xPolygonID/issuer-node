import { Button } from "antd";
import { useNavigate } from "react-router-dom";

import { ReactComponent as IconBack } from "src/assets/icons/arrow-narrow-left.svg";

export function BackButton() {
  const navigate = useNavigate();

  return (
    <Button icon={<IconBack />} onClick={() => navigate(-1)} style={{ paddingLeft: 0 }} type="link">
      Back
    </Button>
  );
}
