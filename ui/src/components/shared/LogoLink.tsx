import { Link } from "react-router-dom";

import { ReactComponent as IconLogo } from "src/assets/polygonid-logo.svg";
import { ROOT_PATH } from "src/utils/constants";

export function LogoLink() {
  return (
    <Link to={ROOT_PATH}>
      <IconLogo />
    </Link>
  );
}
